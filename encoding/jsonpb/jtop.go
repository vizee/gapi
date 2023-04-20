package jsonpb

import (
	"encoding/base64"
	"errors"
	"io"
	"math"
	"strconv"

	"github.com/vizee/gapi/encoding/jsonlit"
	"github.com/vizee/gapi/encoding/proto"
	"github.com/vizee/gapi/metadata"
)

type JsonIter = jsonlit.Iter[[]byte]

var (
	ErrUnexpectedToken = errors.New("unexpected token")
	ErrTypeMismatch    = errors.New("field type mismatch")
)

func transJsonRepeatedMessage(p *proto.Encoder, j *JsonIter, field *metadata.Field) error {
	var buf proto.Encoder
	for !j.EOF() {
		tok, _ := j.Next()
		switch tok {
		case jsonlit.ArrayClose:
			return nil
		case jsonlit.Comma:
		case jsonlit.Object:
			buf.Clear()
			err := transJsonObject(&buf, j, field.Ref)
			if err != nil {
				return err
			}
			p.EmitBytes(field.Tag, buf.Bytes())
		case jsonlit.Null:
			// null 会表达为一个空对象占位
			p.EmitBytes(field.Tag, nil)
		default:
			return ErrUnexpectedToken
		}
	}
	return io.ErrUnexpectedEOF
}

func walkJsonArray(j *JsonIter, expect jsonlit.Kind, f func([]byte) error) error {
	for !j.EOF() {
		tok, s := j.Next()
		switch tok {
		case jsonlit.ArrayClose:
			return nil
		case jsonlit.Comma:
		case expect:
			err := f(s)
			if err != nil {
				return err
			}
		default:
			return ErrUnexpectedToken
		}
	}
	return io.ErrUnexpectedEOF
}

func transJsonArrayField(p *proto.Encoder, j *JsonIter, field *metadata.Field) error {
	switch field.Kind {
	case metadata.MessageKind:
		return transJsonRepeatedMessage(p, j, field)
	case metadata.BytesKind:
		// 暂不允许 null 转到 bytes
		err := walkJsonArray(j, jsonlit.String, func(s []byte) error {
			return transJsonBytes(p, field.Tag, false, s)
		})
		if err != nil {
			return err
		}
	case metadata.StringKind:
		err := walkJsonArray(j, jsonlit.String, func(s []byte) error {
			return transJsonString(p, field.Tag, false, s)
		})
		if err != nil {
			return err
		}
	default:
		var (
			packed proto.Encoder
			err    error
		)
		switch field.Kind {
		case metadata.DoubleKind:
			err = walkJsonArray(j, jsonlit.Number, func(s []byte) error {
				x, err := strconv.ParseFloat(bytesView(s), 64)
				if err != nil {
					return err
				}
				packed.WriteFixed64(math.Float64bits(x))
				return nil
			})
		case metadata.FloatKind:
			err = walkJsonArray(j, jsonlit.Number, func(s []byte) error {
				x, err := strconv.ParseFloat(bytesView(s), 32)
				if err != nil {
					return err
				}
				packed.WriteFixed32(math.Float32bits(float32(x)))
				return nil
			})
		case metadata.Int32Kind:
			err = walkJsonArray(j, jsonlit.Number, func(s []byte) error {
				x, err := strconv.ParseInt(bytesView(s), 10, 32)
				if err != nil {
					return err
				}
				packed.WriteVarint(uint64(x))
				return nil
			})
		case metadata.Int64Kind:
			err = walkJsonArray(j, jsonlit.Number, func(s []byte) error {
				x, err := strconv.ParseInt(bytesView(s), 10, 64)
				if err != nil {
					return err
				}
				packed.WriteVarint(uint64(x))
				return nil
			})
		case metadata.Uint32Kind:
			err = walkJsonArray(j, jsonlit.Number, func(s []byte) error {
				x, err := strconv.ParseUint(bytesView(s), 10, 32)
				if err != nil {
					return err
				}
				packed.WriteVarint(x)
				return nil
			})
		case metadata.Uint64Kind:
			err = walkJsonArray(j, jsonlit.Number, func(s []byte) error {
				x, err := strconv.ParseUint(bytesView(s), 10, 64)
				if err != nil {
					return err
				}
				packed.WriteVarint(x)
				return nil
			})
		case metadata.Sint32Kind:
			err = walkJsonArray(j, jsonlit.Number, func(s []byte) error {
				x, err := strconv.ParseInt(bytesView(s), 10, 32)
				if err != nil {
					return err
				}
				packed.WriteZigzag(x)
				return nil
			})
		case metadata.Sint64Kind:
			err = walkJsonArray(j, jsonlit.Number, func(s []byte) error {
				x, err := strconv.ParseInt(bytesView(s), 10, 64)
				if err != nil {
					return err
				}
				packed.WriteZigzag(x)
				return nil
			})
		case metadata.Fixed32Kind:
			err = walkJsonArray(j, jsonlit.Number, func(s []byte) error {
				x, err := strconv.ParseUint(bytesView(s), 10, 32)
				if err != nil {
					return err
				}
				packed.WriteFixed32(uint32(x))
				return nil
			})
		case metadata.Fixed64Kind:
			err = walkJsonArray(j, jsonlit.Number, func(s []byte) error {
				x, err := strconv.ParseUint(bytesView(s), 10, 64)
				if err != nil {
					return err
				}
				packed.WriteFixed64(x)
				return nil
			})
		case metadata.Sfixed32Kind:
			err = walkJsonArray(j, jsonlit.Number, func(s []byte) error {
				x, err := strconv.ParseInt(bytesView(s), 10, 32)
				if err != nil {
					return err
				}
				packed.WriteFixed32(uint32(x))
				return nil
			})
		case metadata.Sfixed64Kind:
			err = walkJsonArray(j, jsonlit.Number, func(s []byte) error {
				x, err := strconv.ParseInt(bytesView(s), 10, 64)
				if err != nil {
					return err
				}
				packed.WriteFixed64(uint64(x))
				return nil
			})
		case metadata.BoolKind:
			err = walkJsonArray(j, jsonlit.Bool, func(s []byte) error {
				var x uint64
				if len(s) == 4 {
					x = 1
				} else {
					x = 0
				}
				packed.WriteVarint(x)
				return nil
			})
		default:
			err = ErrTypeMismatch
		}
		if err != nil {
			return err
		}
		if packed.Len() != 0 {
			p.EmitBytes(field.Tag, packed.Bytes())
		}
	}
	return nil
}

func transJsonToMap(p *proto.Encoder, j *JsonIter, tag uint32, entry *metadata.Message) error {
	// assert(len(entry.Fields) == 2)
	keyField, valueField := &entry.Fields[0], &entry.Fields[1]
	var buf proto.Encoder
	expectValue := false
	for !j.EOF() {
		lead, s := j.Next()
		switch lead {
		case jsonlit.ObjectClose:
			if expectValue {
				return ErrUnexpectedToken
			}
			return nil
		case jsonlit.Comma, jsonlit.Colon:
			// 忽略语法检查
			continue
		default:
			if expectValue {
				// NOTE: transJsonField 会跳过 0 值字段，导致结果比 proto.Marshal 的结果字节数更少，但不影响反序列化结果
				err := transJsonField(&buf, j, valueField, lead, s)
				if err != nil {
					return err
				}
				if buf.Len() != 0 {
					p.EmitBytes(tag, buf.Bytes())
				}
				expectValue = false
			} else if lead == jsonlit.String {
				buf.Clear()
				if keyField.Kind == metadata.StringKind {
					err := transJsonString(&buf, keyField.Tag, true, s)
					if err != nil {
						return err
					}
				} else if metadata.IsNumericKind(keyField.Kind) {
					// 允许把 json key 转为将数值类型的 map key
					err := transJsonNumeric(&buf, keyField.Tag, keyField.Kind, s[1:len(s)-1])
					if err != nil {
						return err
					}
				} else {
					return ErrTypeMismatch
				}
				expectValue = true
			} else {
				return ErrUnexpectedToken
			}
		}
	}
	return io.ErrUnexpectedEOF
}

func transJsonNumeric(p *proto.Encoder, tag uint32, kind metadata.Kind, s []byte) error {
	if !metadata.IsNumericKind(kind) {
		return ErrTypeMismatch
	}
	// 提前检查 0 值
	if len(s) == 1 && s[0] == '0' {
		return nil
	}
	switch kind {
	case metadata.DoubleKind:
		x, err := strconv.ParseFloat(bytesView(s), 64)
		if err != nil {
			return err
		}
		p.EmitFixed64(tag, math.Float64bits(x))
	case metadata.FloatKind:
		x, err := strconv.ParseFloat(bytesView(s), 32)
		if err != nil {
			return err
		}
		p.EmitFixed32(tag, math.Float32bits(float32(x)))
	case metadata.Int32Kind:
		x, err := strconv.ParseInt(bytesView(s), 10, 32)
		if err != nil {
			return err
		}
		p.EmitVarint(tag, uint64(x))
	case metadata.Int64Kind:
		x, err := strconv.ParseInt(bytesView(s), 10, 64)
		if err != nil {
			return err
		}
		p.EmitVarint(tag, uint64(x))
	case metadata.Uint32Kind:
		x, err := strconv.ParseUint(bytesView(s), 10, 32)
		if err != nil {
			return err
		}
		p.EmitVarint(tag, uint64(x))
	case metadata.Uint64Kind:
		x, err := strconv.ParseUint(bytesView(s), 10, 64)
		if err != nil {
			return err
		}
		p.EmitVarint(tag, x)
	case metadata.Sint32Kind:
		x, err := strconv.ParseInt(bytesView(s), 10, 32)
		if err != nil {
			return err
		}
		p.EmitZigzag(tag, x)
	case metadata.Sint64Kind:
		x, err := strconv.ParseInt(bytesView(s), 10, 64)
		if err != nil {
			return err
		}
		p.EmitZigzag(tag, x)
	case metadata.Fixed32Kind:
		x, err := strconv.ParseUint(bytesView(s), 10, 32)
		if err != nil {
			return err
		}
		p.EmitFixed32(tag, uint32(x))
	case metadata.Fixed64Kind:
		x, err := strconv.ParseUint(bytesView(s), 10, 64)
		if err != nil {
			return err
		}
		p.EmitFixed64(tag, x)
	case metadata.Sfixed32Kind:
		x, err := strconv.ParseInt(bytesView(s), 10, 32)
		if err != nil {
			return err
		}
		p.EmitFixed32(tag, uint32(x))
	case metadata.Sfixed64Kind:
		x, err := strconv.ParseInt(bytesView(s), 10, 64)
		if err != nil {
			return err
		}
		p.EmitFixed64(tag, uint64(x))
	}
	return nil
}

func transJsonString(p *proto.Encoder, tag uint32, omitEmpty bool, s []byte) error {
	if len(s) == 2 && omitEmpty {
		return nil
	}
	z := make([]byte, 0, len(s)-2)
	z, ok := jsonlit.UnescapeString(z, s[1:len(s)-1])
	if !ok {
		return errors.New("unescape malformed string")
	}
	p.EmitBytes(tag, z)
	return nil
}

func transJsonBytes(p *proto.Encoder, tag uint32, omitEmpty bool, s []byte) error {
	if len(s) == 2 && omitEmpty {
		return nil
	}
	z := make([]byte, base64.StdEncoding.DecodedLen(len(s)-2))
	n, err := base64.StdEncoding.Decode(z, s[1:len(s)-1])
	if err != nil {
		return err
	}
	p.EmitBytes(tag, z[:n])
	return nil
}

func transJsonField(p *proto.Encoder, j *JsonIter, field *metadata.Field, lead jsonlit.Kind, s []byte) error {
	switch lead {
	case jsonlit.String:
		switch field.Kind {
		case metadata.BytesKind:
			return transJsonBytes(p, field.Tag, true, s)
		case metadata.StringKind:
			return transJsonString(p, field.Tag, true, s)
		default:
			return ErrTypeMismatch
		}
	case jsonlit.Number:
		return transJsonNumeric(p, field.Tag, field.Kind, s)
	case jsonlit.Bool:
		if field.Kind == metadata.BoolKind {
			if len(s) == 4 {
				p.EmitVarint(field.Tag, 1)
			}
			return nil
		} else {
			return ErrTypeMismatch
		}
	case jsonlit.Null:
		// 忽略所有 null
		return nil
	case jsonlit.Object:
		switch field.Kind {
		case metadata.MessageKind:
			var buf proto.Encoder
			err := transJsonObject(&buf, j, field.Ref)
			if err != nil {
				return err
			}
			if buf.Len() != 0 {
				p.EmitBytes(field.Tag, buf.Bytes())
			}
			return nil
		case metadata.MapKind:
			return transJsonToMap(p, j, field.Tag, field.Ref)
		default:
			return ErrTypeMismatch
		}
	case jsonlit.Array:
		if field.Repeated {
			return transJsonArrayField(p, j, field)
		}
		return ErrTypeMismatch
	}
	return ErrUnexpectedToken
}

func skipJsonValue(j *JsonIter, lead jsonlit.Kind) error {
	switch lead {
	case jsonlit.Null, jsonlit.Bool, jsonlit.Number, jsonlit.String:
		return nil
	case jsonlit.Object:
		for !j.EOF() {
			tok, _ := j.Next()
			switch tok {
			case jsonlit.ObjectClose:
				return nil
			case jsonlit.Comma, jsonlit.Colon:
			default:
				err := skipJsonValue(j, tok)
				if err != nil {
					return err
				}
			}
		}
	case jsonlit.Array:
		for !j.EOF() {
			tok, _ := j.Next()
			switch tok {
			case jsonlit.ArrayClose:
				return nil
			case jsonlit.Comma:
			default:
				err := skipJsonValue(j, tok)
				if err != nil {
					return err
				}
			}
		}
		return io.ErrUnexpectedEOF
	}
	return ErrUnexpectedToken
}

func transJsonObject(p *proto.Encoder, j *JsonIter, msg *metadata.Message) error {
	var key []byte
	for !j.EOF() {
		lead, s := j.Next()
		switch lead {
		case jsonlit.ObjectClose:
			if len(key) == 0 {
				return nil
			}
			return io.ErrUnexpectedEOF
		case jsonlit.Comma, jsonlit.Colon:
			// 忽略语法检查
			continue
		default:
			if len(key) != 0 {
				// 暂不转义 key
				field := msg.FieldByName(bytesView(key[1 : len(key)-1]))
				if field != nil {
					err := transJsonField(p, j, field, lead, s)
					if err != nil {
						return err
					}
				} else {
					err := skipJsonValue(j, lead)
					if err != nil {
						return err
					}
				}
				key = nil
			} else if lead == jsonlit.String {
				key = s
			} else {
				return ErrUnexpectedToken
			}
		}
	}
	return io.ErrUnexpectedEOF
}

// Jtop 通过 JsonIter 解析 JSON，并且根据 msg 将 JSON 内容转译到 protobuf 二进制。
// 注意，受限于 metadata 可表达的结构和一些取舍，对 JSON 的解析并不按照 JSON 标准。
func Jtop(p *proto.Encoder, j *JsonIter, msg *metadata.Message) error {
	tok, _ := j.Next()
	switch tok {
	case jsonlit.Object:
		return transJsonObject(p, j, msg)
	case jsonlit.EOF:
		return io.ErrUnexpectedEOF
	}
	return ErrUnexpectedToken
}

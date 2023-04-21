package jsonpb

import (
	"bytes"
	"encoding/base64"
	"errors"
	"math"
	"strconv"

	"github.com/vizee/gapi/encoding/jsonlit"
	"github.com/vizee/gapi/encoding/proto"
	"github.com/vizee/gapi/metadata"
	"google.golang.org/protobuf/encoding/protowire"
)

type protoValue struct {
	x uint64
	s []byte
}

func readProtoValue(p *proto.Decoder, wire protowire.Type) (val protoValue, e int) {
	switch wire {
	case protowire.VarintType:
		val.x, e = p.ReadVarint()
	case protowire.Fixed32Type:
		var t uint32
		t, e = p.ReadFixed32()
		val.x = uint64(t)
	case protowire.Fixed64Type:
		val.x, e = p.ReadFixed64()
	case protowire.BytesType:
		val.s, e = p.ReadBytes()
	default:
		e = -100
	}
	return
}

var (
	ErrInvalidWireType = errors.New("invalid wire type")
)

var defaultValues = [...]string{
	metadata.DoubleKind:   `0`,
	metadata.FloatKind:    `0`,
	metadata.Int32Kind:    `0`,
	metadata.Int64Kind:    `0`,
	metadata.Uint32Kind:   `0`,
	metadata.Uint64Kind:   `0`,
	metadata.Sint32Kind:   `0`,
	metadata.Sint64Kind:   `0`,
	metadata.Fixed32Kind:  `0`,
	metadata.Fixed64Kind:  `0`,
	metadata.Sfixed32Kind: `0`,
	metadata.Sfixed64Kind: `0`,
	metadata.BoolKind:     `false`,
	metadata.StringKind:   `""`,
	metadata.BytesKind:    `""`,
	metadata.MapKind:      `{}`,
	metadata.MessageKind:  `{}`,
}

func writeDefaultValue(w *bytes.Buffer, repeated bool, kind metadata.Kind) (err error) {
	if repeated {
		_, err = w.Write(asBytes("[]"))
	} else {
		_, err = w.Write(asBytes(defaultValues[kind]))
	}
	return
}

func transProtoMap(w *bytes.Buffer, p *proto.Decoder, field *metadata.Field, s []byte) error {
	err := w.WriteByte('{')
	if err != nil {
		return err
	}

	keyField, valueField := field.Ref.FieldByTag(1), field.Ref.FieldByTag(2)
	// assert(keyField != nil && valueField != nil)
	keyWire := getFieldWireType(keyField)
	valueWire := getFieldWireType(valueField)
	// 暂不检查 keyField.Kind

	more := false
	for {
		if !more {
			more = true
		} else {
			err := w.WriteByte(',')
			if err != nil {
				return err
			}
		}

		// 上下文比较复杂，直接嵌套逻辑读取 KV

		var values [2]protoValue
		assigned := 0
		dec := proto.NewDecoder(s)
		for !dec.EOF() && assigned != 3 {
			tag, wire, e := dec.ReadTag()
			if e < 0 {
				return protowire.ParseError(e)
			}
			val, e := readProtoValue(dec, wire)
			if e < 0 {
				return protowire.ParseError(e)
			}
			switch tag {
			case 1:
				if wire != keyWire {
					return ErrInvalidWireType
				}
				values[0] = val
				assigned |= 1
			case 2:
				if wire != valueWire {
					return ErrInvalidWireType
				}
				values[1] = val
				assigned |= 2
			}
		}

		if assigned&1 != 0 {
			if keyField.Kind == metadata.StringKind {
				err := transProtoString(w, values[0].s)
				if err != nil {
					return err
				}
			} else {
				err := w.WriteByte('"')
				if err != nil {
					return err
				}
				err = transProtoSimpleValue(w, field.Kind, values[0].x)
				if err != nil {
					return err
				}
				err = w.WriteByte('"')
				if err != nil {
					return err
				}
			}
		} else {
			_, err := w.Write(asBytes(`""`))
			if err != nil {
				return err
			}
		}

		err := w.WriteByte(':')
		if err != nil {
			return err
		}

		if assigned&2 != 0 {
			switch valueField.Kind {
			case metadata.StringKind:
				err = transProtoString(w, values[1].s)
			case metadata.BytesKind:
				err = transProtoBytes(w, values[1].s)
			case metadata.MessageKind:
				err = transProtoMessage(w, proto.NewDecoder(values[1].s), valueField.Ref)
			default:
				err = transProtoSimpleValue(w, field.Kind, values[1].x)
			}
		} else {
			err := writeDefaultValue(w, valueField.Repeated, valueField.Kind)
			if err != nil {
				return err
			}
		}

		if p.EOF() {
			break
		}
		tag, wire, e := p.PeekTag()
		if e < 0 {
			return protowire.ParseError(e)
		}
		if tag != field.Tag {
			break
		}
		if wire != protowire.BytesType {
			return ErrInvalidWireType
		}
		p.ReadVarint() // consume tag
		s, e = p.ReadBytes()
		if e < 0 {
			return protowire.ParseError(e)
		}
	}

	return w.WriteByte('}')
}

func transProtoRepeatedBytes(w *bytes.Buffer, p *proto.Decoder, field *metadata.Field, s []byte) error {
	err := w.WriteByte('[')
	if err != nil {
		return err
	}

	more := false
	for {
		if !more {
			more = true
		} else {
			err := w.WriteByte(',')
			if err != nil {
				return err
			}
		}

		switch field.Kind {
		case metadata.StringKind:
			err := transProtoString(w, s)
			if err != nil {
				return err
			}
		case metadata.BytesKind:
			err := transProtoBytes(w, s)
			if err != nil {
				return err
			}
		case metadata.MessageKind:
			err := transProtoMessage(w, proto.NewDecoder(s), field.Ref)
			if err != nil {
				return err
			}
		}

		if p.EOF() {
			break
		}
		tag, wire, e := p.PeekTag()
		if e < 0 {
			return protowire.ParseError(e)
		}
		if tag != field.Tag {
			break
		}
		if wire != protowire.BytesType {
			return ErrInvalidWireType
		}
		p.ReadVarint() // consume tag
		s, e = p.ReadBytes()
		if e < 0 {
			return protowire.ParseError(e)
		}
	}

	return w.WriteByte(']')
}

func transProtoPackedArray(w *bytes.Buffer, p *proto.Decoder, field *metadata.Field) error {
	err := w.WriteByte('[')
	if err != nil {
		return err
	}

	wire := getFieldWireType(field)
	more := false
	for !p.EOF() {
		if !more {
			more = true
		} else {
			err := w.WriteByte(',')
			if err != nil {
				return err
			}
		}
		val, e := readProtoValue(p, wire)
		if e < 0 {
			return protowire.ParseError(e)
		}
		err := transProtoSimpleValue(w, field.Kind, val.x)
		if err != nil {
			return err
		}
	}

	return w.WriteByte(']')
}

func transProtoBytes(w *bytes.Buffer, s []byte) error {
	err := w.WriteByte('"')
	if err != nil {
		return err
	}
	buf := make([]byte, base64.StdEncoding.EncodedLen(len(s)))
	base64.StdEncoding.Encode(buf, s)
	_, err = w.Write(buf)
	if err != nil {
		return err
	}
	return w.WriteByte('"')
}

func transProtoString(w *bytes.Buffer, s []byte) error {
	err := w.WriteByte('"')
	if err != nil {
		return err
	}
	_, err = w.Write(jsonlit.EscapeString(make([]byte, 0, len(s)), s))
	if err != nil {
		return err
	}
	return w.WriteByte('"')
}

func transProtoSimpleValue(w *bytes.Buffer, kind metadata.Kind, x uint64) error {
	var pre [32]byte
	buf := pre[:0]
	switch kind {
	case metadata.DoubleKind:
		buf = strconv.AppendFloat(buf, math.Float64frombits(x), 'f', -1, 64)
	case metadata.FloatKind:
		buf = strconv.AppendFloat(buf, float64(math.Float32frombits(uint32(x))), 'f', -1, 32)
	case metadata.Int32Kind, metadata.Int64Kind, metadata.Sfixed32Kind, metadata.Sfixed64Kind:
		buf = strconv.AppendInt(buf, int64(x), 10)
	case metadata.Uint32Kind, metadata.Uint64Kind, metadata.Fixed32Kind, metadata.Fixed64Kind:
		buf = strconv.AppendUint(buf, x, 10)
	case metadata.Sint32Kind:
		buf = strconv.AppendInt(buf, int64(protowire.DecodeZigZag(uint64(int64(int32(x))))), 10)
	case metadata.Sint64Kind:
		buf = strconv.AppendInt(buf, int64(protowire.DecodeZigZag(x)), 10)
	case metadata.BoolKind:
		buf = strconv.AppendBool(buf, x != 0)
	}
	_, err := w.Write(buf)
	return err
}

var wireTypeOfKind = [...]protowire.Type{
	metadata.DoubleKind:   protowire.Fixed64Type,
	metadata.FloatKind:    protowire.Fixed32Type,
	metadata.Int32Kind:    protowire.VarintType,
	metadata.Int64Kind:    protowire.VarintType,
	metadata.Uint32Kind:   protowire.VarintType,
	metadata.Uint64Kind:   protowire.VarintType,
	metadata.Sint32Kind:   protowire.VarintType,
	metadata.Sint64Kind:   protowire.VarintType,
	metadata.Fixed32Kind:  protowire.Fixed32Type,
	metadata.Fixed64Kind:  protowire.Fixed64Type,
	metadata.Sfixed32Kind: protowire.Fixed32Type,
	metadata.Sfixed64Kind: protowire.Fixed64Type,
	metadata.BoolKind:     protowire.VarintType,
	// metadata.StringKind:   protowire.BytesType,
	// metadata.BytesKind:    protowire.BytesType,
	// metadata.MapKind:      protowire.BytesType,
	// metadata.MessageKind:  protowire.BytesType,
}

func getFieldWireType(field *metadata.Field) protowire.Type {
	if field.Repeated {
		// 如果字段设置 repeated，那么值应该是 packed/string/bytes/message，所以 wire 一定是 BytesType
		return protowire.BytesType
	} else if int(field.Kind) < len(wireTypeOfKind) {
		return wireTypeOfKind[field.Kind]
	}
	return protowire.BytesType
}

func transProtoMessage(w *bytes.Buffer, p *proto.Decoder, msg *metadata.Message) error {
	err := w.WriteByte('{')
	if err != nil {
		return err
	}

	more := false
	for !p.EOF() {
		tag, wire, e := p.ReadTag()
		if e < 0 {
			return protowire.ParseError(e)
		}

		val, e := readProtoValue(p, wire)
		if e < 0 {
			return protowire.ParseError(e)
		}

		field := msg.FieldByTag(tag)
		if field == nil {
			continue
		}
		expectedWire := getFieldWireType(field)
		if expectedWire != wire {
			return ErrInvalidWireType
		}

		if !more {
			more = true
		} else {
			err = w.WriteByte(',')
			if err != nil {
				return err
			}
		}
		err = w.WriteByte('"')
		if err != nil {
			return err
		}
		_, err := w.Write(asBytes(field.Name))
		if err != nil {
			return err
		}
		_, err = w.Write(asBytes("\":"))
		if err != nil {
			return err
		}

		if field.Repeated {
			switch field.Kind {
			case metadata.StringKind, metadata.BytesKind, metadata.MessageKind:
				err = transProtoRepeatedBytes(w, p, field, val.s)
			default:
				err = transProtoPackedArray(w, p, field)
			}
		} else if field.Kind == metadata.MapKind {
			err = transProtoMap(w, p, field, val.s)
		} else {
			switch field.Kind {
			case metadata.StringKind:
				err = transProtoString(w, val.s)
			case metadata.BytesKind:
				err = transProtoBytes(w, val.s)
			case metadata.MessageKind:
				err = transProtoMessage(w, proto.NewDecoder(val.s), field.Ref)
			default:
				err = transProtoSimpleValue(w, field.Kind, val.x)
			}
		}
		if err != nil {
			return err
		}
	}

	return w.WriteByte('}')
}

func TranscodeToJson(w *bytes.Buffer, p *proto.Decoder, msg *metadata.Message) error {
	return transProtoMessage(w, p, msg)
}

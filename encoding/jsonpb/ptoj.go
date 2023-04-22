package jsonpb

import (
	"encoding/base64"
	"errors"
	"math"
	"strconv"

	"github.com/vizee/gapi/encoding/proto"
	"github.com/vizee/gapi/metadata"
	"google.golang.org/protobuf/encoding/protowire"
)

var (
	ErrInvalidWireType = errors.New("invalid wire type")
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

func writeDefaultValue(j *JsonBuilder, repeated bool, kind metadata.Kind) {
	if repeated {
		j.AppendString("[]")
	} else {
		j.AppendString(defaultValues[kind])
	}
}

func transProtoMap(j *JsonBuilder, p *proto.Decoder, field *metadata.Field, s []byte) error {
	j.AppendByte('{')

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
			j.AppendByte(',')
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
				err := transProtoString(j, values[0].s)
				if err != nil {
					return err
				}
			} else {
				j.AppendByte('"')
				err := transProtoSimpleValue(j, field.Kind, values[0].x)
				if err != nil {
					return err
				}
				j.AppendByte('"')
			}
		} else {
			j.AppendString(`""`)
		}

		j.AppendByte(':')

		if assigned&2 != 0 {
			var err error
			switch valueField.Kind {
			case metadata.StringKind:
				err = transProtoString(j, values[1].s)
			case metadata.BytesKind:
				err = transProtoBytes(j, values[1].s)
			case metadata.MessageKind:
				err = transProtoMessage(j, proto.NewDecoder(values[1].s), valueField.Ref)
			default:
				err = transProtoSimpleValue(j, field.Kind, values[1].x)
			}
			if err != nil {
				return err
			}
		} else {
			writeDefaultValue(j, valueField.Repeated, valueField.Kind)
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

	j.AppendByte('}')
	return nil
}

func transProtoRepeatedBytes(j *JsonBuilder, p *proto.Decoder, field *metadata.Field, s []byte) error {
	j.AppendByte('[')

	more := false
	for {
		if !more {
			more = true
		} else {
			j.AppendByte(',')
		}

		switch field.Kind {
		case metadata.StringKind:
			err := transProtoString(j, s)
			if err != nil {
				return err
			}
		case metadata.BytesKind:
			err := transProtoBytes(j, s)
			if err != nil {
				return err
			}
		case metadata.MessageKind:
			err := transProtoMessage(j, proto.NewDecoder(s), field.Ref)
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

	j.AppendByte(']')
	return nil
}

func transProtoPackedArray(j *JsonBuilder, p *proto.Decoder, field *metadata.Field) error {
	j.AppendByte('[')

	wire := getFieldWireType(field)
	more := false
	for !p.EOF() {
		if !more {
			more = true
		} else {
			j.AppendByte(',')
		}
		val, e := readProtoValue(p, wire)
		if e < 0 {
			return protowire.ParseError(e)
		}
		err := transProtoSimpleValue(j, field.Kind, val.x)
		if err != nil {
			return err
		}
	}

	j.AppendByte(']')
	return nil
}

func transProtoBytes(j *JsonBuilder, s []byte) error {
	j.AppendByte('"')
	n := base64.StdEncoding.EncodedLen(len(s))
	j.Reserve(n)
	m := len(j.buf)
	d := j.buf[m : m+n]
	base64.StdEncoding.Encode(d, s)
	j.buf = j.buf[:m+n]
	j.AppendByte('"')
	return nil
}

func transProtoString(j *JsonBuilder, s []byte) error {
	j.AppendByte('"')
	j.AppendEscapedString(asString(s))
	j.AppendByte('"')
	return nil
}

func transProtoSimpleValue(j *JsonBuilder, kind metadata.Kind, x uint64) error {
	switch kind {
	case metadata.DoubleKind:
		j.buf = strconv.AppendFloat(j.buf, math.Float64frombits(x), 'f', -1, 64)
	case metadata.FloatKind:
		j.buf = strconv.AppendFloat(j.buf, float64(math.Float32frombits(uint32(x))), 'f', -1, 32)
	case metadata.Int32Kind, metadata.Int64Kind, metadata.Sfixed32Kind, metadata.Sfixed64Kind:
		j.buf = strconv.AppendInt(j.buf, int64(x), 10)
	case metadata.Uint32Kind, metadata.Uint64Kind, metadata.Fixed32Kind, metadata.Fixed64Kind:
		j.buf = strconv.AppendUint(j.buf, x, 10)
	case metadata.Sint32Kind:
		j.buf = strconv.AppendInt(j.buf, int64(protowire.DecodeZigZag(uint64(int64(int32(x))))), 10)
	case metadata.Sint64Kind:
		j.buf = strconv.AppendInt(j.buf, int64(protowire.DecodeZigZag(x)), 10)
	case metadata.BoolKind:
		if x != 0 {
			j.AppendString("true")
		} else {
			j.AppendString("false")
		}
	}
	return nil
}

func transProtoMessage(j *JsonBuilder, p *proto.Decoder, msg *metadata.Message) error {
	j.AppendByte('{')

	var pre [32]bool

	emitted := pre[:0]

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

		fieldIdx := msg.FieldIndexByTag(tag)
		if fieldIdx < 0 {
			continue
		}
		field := &msg.Fields[fieldIdx]
		expectedWire := getFieldWireType(field)
		if expectedWire != wire {
			return ErrInvalidWireType
		}

		if emitted[fieldIdx] {
			continue
		}

		if !more {
			more = true
		} else {
			j.AppendByte(',')
		}
		j.AppendByte('"')
		j.AppendString(field.Name)
		j.AppendByte('"')
		j.AppendByte(':')

		var err error
		if field.Repeated {
			switch field.Kind {
			case metadata.StringKind, metadata.BytesKind, metadata.MessageKind:
				err = transProtoRepeatedBytes(j, p, field, val.s)
			default:
				err = transProtoPackedArray(j, p, field)
			}
		} else if field.Kind == metadata.MapKind {
			err = transProtoMap(j, p, field, val.s)
		} else {
			switch field.Kind {
			case metadata.StringKind:
				err = transProtoString(j, val.s)
			case metadata.BytesKind:
				err = transProtoBytes(j, val.s)
			case metadata.MessageKind:
				err = transProtoMessage(j, proto.NewDecoder(val.s), field.Ref)
			default:
				err = transProtoSimpleValue(j, field.Kind, val.x)
			}
		}
		if err != nil {
			return err
		}

		emitted[fieldIdx] = true
	}

	for i := range msg.Fields {
		field := &msg.Fields[i]
		if emitted[i] || field.OmitEmpty {
			continue
		}
		if !more {
			more = true
		} else {
			j.AppendByte(',')
		}
		j.AppendByte('"')
		j.AppendString(field.Name)
		j.AppendByte('"')
		j.AppendByte(':')
		writeDefaultValue(j, field.Repeated, field.Kind)
	}

	j.AppendByte('}')
	return nil
}

func TranscodeToJson(j *JsonBuilder, p *proto.Decoder, msg *metadata.Message) error {
	return transProtoMessage(j, p, msg)
}

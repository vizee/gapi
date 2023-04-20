package jsonpb

import "github.com/vizee/gapi/metadata"

func getTestSimpleMessage() *metadata.Message {
	return metadata.NewMessage("Simple", []metadata.Field{
		{Name: "name", Tag: 1, Kind: metadata.StringKind},
		{Name: "age", Tag: 2, Kind: metadata.Int32Kind},
		{Name: "male", Tag: 3, Kind: metadata.BoolKind},
	}, true, true)
}

func getTestMapEntry(keyKind metadata.Kind, valueKind metadata.Kind, valueRef *metadata.Message) *metadata.Message {
	return metadata.NewMessage("", []metadata.Field{
		0: {Tag: 1, Kind: keyKind},
		1: {Tag: 2, Kind: valueKind, Ref: valueRef},
	}, true, true)
}

func getTestComplexMessage() *metadata.Message {
	return metadata.NewMessage("Complex", []metadata.Field{
		{Name: "fdouble", Kind: metadata.DoubleKind, Tag: 1},
		{Name: "ffloat", Kind: metadata.FloatKind, Tag: 2},
		{Name: "fint32", Kind: metadata.Int32Kind, Tag: 3},
		{Name: "fint64", Kind: metadata.Int64Kind, Tag: 4},
		{Name: "fuint32", Kind: metadata.Uint32Kind, Tag: 5},
		{Name: "fuint64", Kind: metadata.Uint64Kind, Tag: 6},
		{Name: "fsint32", Kind: metadata.Sint32Kind, Tag: 7},
		{Name: "fsint64", Kind: metadata.Sint64Kind, Tag: 8},
		{Name: "ffixed32", Kind: metadata.Fixed32Kind, Tag: 9},
		{Name: "ffixed64", Kind: metadata.Fixed64Kind, Tag: 10},
		{Name: "fsfixed32", Kind: metadata.Sfixed32Kind, Tag: 11},
		{Name: "fsfixed64", Kind: metadata.Sfixed64Kind, Tag: 12},
		{Name: "fbool", Kind: metadata.BoolKind, Tag: 13},
		{Name: "fstring", Kind: metadata.StringKind, Tag: 14},
		{Name: "fbytes", Kind: metadata.BytesKind, Tag: 15},
		{Name: "fmap1", Kind: metadata.MapKind, Tag: 16, Ref: getTestMapEntry(metadata.StringKind, metadata.Int32Kind, nil)},
		{Name: "fmap2", Kind: metadata.MapKind, Tag: 17, Ref: getTestMapEntry(metadata.StringKind, metadata.MessageKind, getTestSimpleMessage())},
		{Name: "fsubmsg", Kind: metadata.MessageKind, Tag: 18, Ref: getTestSimpleMessage()},
		{Name: "fint32s", Kind: metadata.Int32Kind, Tag: 19, Repeated: true},
		{Name: "fitems", Kind: metadata.MessageKind, Tag: 20, Repeated: true, Ref: getTestSimpleMessage()},
	}, true, true)
}

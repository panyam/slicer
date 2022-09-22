package utils

import (
	"errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var marshalOptions = protojson.MarshalOptions{
	Indent:          "  ",
	EmitUnpopulated: true,
	UseProtoNames:   true,
}

func ProtoFormat(msg protoreflect.ProtoMessage) string {
	return marshalOptions.Format(msg)
}

func ProtoToJson(msg protoreflect.ProtoMessage) []byte {
	if msg == nil {
		panic(errors.New("Cnanot marshall nil"))
	}
	jsonData, _ := marshalOptions.Marshal(msg)
	return jsonData
}

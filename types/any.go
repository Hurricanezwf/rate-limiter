package types

import (
	proto "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	any "github.com/golang/protobuf/ptypes/any"
)

func MarshalAny(v proto.Message) (*any.Any, error) {
	any, err := ptypes.MarshalAny(v)
	if err != nil {
		return nil, err
	}
	return any, nil
}

func AnyToString(a *any.Any) (*PB_String, error) {
	var v PB_String
	if err := ptypes.UnmarshalAny(a, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func AnyToBytes(a *any.Any) (*PB_Bytes, error) {
	var v PB_Bytes
	if err := ptypes.UnmarshalAny(a, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func AnyToUint32(a *any.Any) (*PB_Uint32, error) {
	var v PB_Uint32
	if err := ptypes.UnmarshalAny(a, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func AnyToUint64(a *any.Any) (*PB_Uint64, error) {
	var v PB_Uint64
	if err := ptypes.UnmarshalAny(a, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func AnyToInt32(a *any.Any) (*PB_Int32, error) {
	var v PB_Int32
	if err := ptypes.UnmarshalAny(a, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func AnyToInt64(a *any.Any) (*PB_Int64, error) {
	var v PB_Int64
	if err := ptypes.UnmarshalAny(a, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func AnyToFloat32(a *any.Any) (*PB_Float32, error) {
	var v PB_Float32
	if err := ptypes.UnmarshalAny(a, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func AnyToFloat64(a *any.Any) (*PB_Float64, error) {
	var v PB_Float64
	if err := ptypes.UnmarshalAny(a, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func AnyToBool(a *any.Any) (*PB_Bool, error) {
	var v PB_Bool
	if err := ptypes.UnmarshalAny(a, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func AnyToMap(a *any.Any) (*PB_Map, error) {
	var v PB_Map
	if err := ptypes.UnmarshalAny(a, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func AnyToQueue(a *any.Any) (*PB_Queue, error) {
	var v PB_Queue
	if err := ptypes.UnmarshalAny(a, &v); err != nil {
		return nil, err
	}
	return &v, nil
}

func UnmarshalAny(a *any.Any, v proto.Message) error {
	if err := ptypes.UnmarshalAny(a, v); err != nil {
		return err
	}
	return nil
}

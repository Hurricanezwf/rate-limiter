package types

import any "github.com/golang/protobuf/ptypes/any"

func NewMap() *PB_Map {
	return &PB_Map{
		Value: make(map[string]*any.Any),
	}
}

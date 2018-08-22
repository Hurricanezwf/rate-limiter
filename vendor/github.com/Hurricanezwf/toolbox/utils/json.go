package utils

import (
	"errors"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func JsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func JsonUnmarshal(b []byte, v interface{}) error {
	if len(b) <= 0 {
		return errors.New("Empty bytes sequence")
	}
	if v == nil {
		return errors.New("The param `v` is nil")
	}
	return json.Unmarshal(b, v)
}

func JsonMarshal2Str(v interface{}) string {
	if v != nil {
		b, err := json.Marshal(v)
		if err == nil {
			return string(b)
		}
	}
	return ""
}

func JsonMarshal2StrWithIndent(v interface{}) string {
	if v != nil {
		b, err := json.MarshalIndent(v, "", "    ")
		if err == nil {
			return string(b)
		}
	}
	return ""
}

package models

import (
	"encoding/json"
	"reflect"
)

func FromJSON(payload []byte, v Validator) error {
	err := json.Unmarshal(payload, v)
	if err != nil {
		return err
	}
	return v.Validate()
}

func ToJSON(v Validator) ([]byte, error) {
	if isNil(v) {
		return json.Marshal(v)
	}

	if err := v.Validate(); err != nil {
		return []byte{}, err
	}

	return json.Marshal(v)
}

func ToJSONArray(vs ...Validator) ([]byte, error) {
	msgs := make([]*json.RawMessage, len(vs))

	for i, v := range vs {
		msg, err := ToJSON(v)
		if err != nil {
			return nil, err
		}

		msgs[i] = (*json.RawMessage)(&msg)
	}

	return json.Marshal(msgs)
}

func isNil(a interface{}) bool {
	if a == nil {
		return true
	}

	switch reflect.TypeOf(a).Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return reflect.ValueOf(a).IsNil()
	}

	return false
}

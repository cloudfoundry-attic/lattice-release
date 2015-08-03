package models

import (
	"encoding/json"
	"reflect"
)

type Routes map[string]*json.RawMessage

func (r *Routes) protoRoutes() *ProtoRoutes {
	pr := &ProtoRoutes{
		Routes: map[string][]byte{},
	}

	for k, v := range *r {
		pr.Routes[k] = *v
	}

	return pr
}

func (r Routes) Marshal() ([]byte, error) {
	return r.protoRoutes().Marshal()
}

func (r Routes) MarshalTo(data []byte) (n int, err error) {
	return r.protoRoutes().MarshalTo(data)
}

func (r *Routes) Unmarshal(data []byte) error {
	pr := ProtoRoutes{}
	err := pr.Unmarshal(data)
	if err != nil {
		return err
	}

	if *r == nil {
		*r = map[string]*json.RawMessage{}
	}

	for k, v := range pr.Routes {
		value := v
		(*r)[k] = (*json.RawMessage)(&value)
	}

	return nil
}

func (r *Routes) Size() int {
	if r == nil {
		return 0
	}

	return r.protoRoutes().Size()
}

func (r Routes) Equal(other Routes) bool {
	return reflect.DeepEqual(r, other)
}

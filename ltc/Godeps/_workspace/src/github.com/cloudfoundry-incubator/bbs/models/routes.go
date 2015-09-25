package models

import (
	"bytes"
	"encoding/json"
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

func (r *Routes) Marshal() ([]byte, error) {
	return r.protoRoutes().Marshal()
}

func (r *Routes) MarshalTo(data []byte) (n int, err error) {
	return r.protoRoutes().MarshalTo(data)
}

func (r *Routes) Unmarshal(data []byte) error {
	pr := &ProtoRoutes{}
	err := pr.Unmarshal(data)
	if err != nil {
		return err
	}

	if pr.Routes == nil {
		return nil
	}

	routes := map[string]*json.RawMessage{}
	for k, v := range pr.Routes {
		raw := json.RawMessage(v)
		routes[k] = &raw
	}
	*r = routes

	return nil
}

func (r *Routes) Size() int {
	if r == nil {
		return 0
	}

	return r.protoRoutes().Size()
}

func (r *Routes) Equal(other Routes) bool {
	for k, v := range *r {
		if !bytes.Equal(*v, *other[k]) {
			return false
		}
	}
	return true
}

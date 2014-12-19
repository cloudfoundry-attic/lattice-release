package persister

import (
	"encoding/json"
)

func NewFakePersister() Persister {
	return &fakePersister{}
}

type fakePersister struct {
	err     error
	content []byte
}

func NewFakePersisterWithError(err error) Persister {
	return &fakePersister{err, []byte{}}
}

func (f *fakePersister) Load(data interface{}) error {
	if f.err != nil {
		return f.err
	}

	return json.Unmarshal(f.content, data)
}

func (f *fakePersister) Save(data interface{}) error {
	if f.err != nil {
		return f.err
	}

	var err error
	f.content, err = json.Marshal(data)
	return err
}

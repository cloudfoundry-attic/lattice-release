package persister

import (
	"encoding/json"
)

func NewMemPersister() Persister {
	return &memPersister{}
}

type memPersister struct {
	content []byte
}

func (m *memPersister) Load(data interface{}) error {
	return json.Unmarshal(m.content, data)
}

func (m *memPersister) Save(data interface{}) error {
	var err error
	m.content, err = json.Marshal(data)
	return err
}

package persister

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type filePersister struct {
	filePath string
}

func NewFilePersister(filepath string) Persister {
	return &filePersister{filepath}
}

func (f *filePersister) Load(i interface{}) error {
	jsonBytes, err := ioutil.ReadFile(f.filePath)
	if err != nil {
		return nil
	}

	err = json.Unmarshal(jsonBytes, i)
	if err != nil {
		return err
	}

	return nil
}

func (f *filePersister) Save(i interface{}) error {
	err := os.MkdirAll(filepath.Dir(f.filePath), 0700)
	if err != nil {
		return err
	}

	jsonBytes, err := json.Marshal(i)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(f.filePath, jsonBytes, 0700)
	if err != nil {
		return err
	}

	return nil
}

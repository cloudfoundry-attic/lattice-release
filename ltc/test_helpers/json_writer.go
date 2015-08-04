package test_helpers

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

func TempJsonFile(obj interface{}, callback func(string)) error {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	file, err := ioutil.TempFile("", "lattice-json")
	if err != nil {
		return err
	}

	_, err = file.Write(bytes)
	if err != nil {
		return err
	}

	filename, err := filepath.Abs(file.Name())
	if err != nil {
		return err
	}

	callback(filename)

	err = os.Remove(filename)
	if err != nil {
		return err
	}

	return nil
}

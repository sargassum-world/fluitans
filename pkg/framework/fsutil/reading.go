// Package fsutil contains utilities for working with fs.FS filesystems.
package fsutil

import (
	"io/fs"
	"io/ioutil"
)

func ReadFile(filename string, f fs.FS) ([]byte, error) {
	file, err := f.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return data, err
}

func ReadConcatenated(filenames []string, f fs.FS) ([]byte, error) {
	var concatenated []byte
	for _, name := range filenames {
		data, err := ReadFile(name, f)
		if err != nil {
			return nil, err
		}
		concatenated = append(concatenated, data...)
	}

	return concatenated, nil
}

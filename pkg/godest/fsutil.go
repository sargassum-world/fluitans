package godest

import (
	"io/fs"
	"io/ioutil"
)

// Directories

func listDirectories(f fs.FS, filter func(path string) bool) ([]string, error) {
	dirs := []string{}
	err := fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if filter == nil || filter(path) {
			if path == "." {
				dirs = append(dirs, "")
			} else {
				dirs = append(dirs, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dirs, nil
}

func listFiles(f fs.FS, filter func(path string) bool) ([]string, error) {
	files := []string{}
	err := fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filter == nil || filter(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// File Reading

func readFile(filename string, f fs.FS) ([]byte, error) {
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

func readConcatenated(filenames []string, f fs.FS) ([]byte, error) {
	var concatenated []byte
	for _, name := range filenames {
		data, err := readFile(name, f)
		if err != nil {
			return nil, err
		}
		concatenated = append(concatenated, data...)
	}

	return concatenated, nil
}

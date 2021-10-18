package fsutil

import (
	"io/fs"
)

func ListFiles(f fs.FS, filter func(path string) bool) ([]string, error) {
	files := []string{}
	err := fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
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

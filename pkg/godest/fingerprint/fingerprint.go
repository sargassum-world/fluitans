// Package fingerprint contains utilities for fingerprinting files from
// fs.FS filesystems, for Etag-based caching of template-based resources.
package fingerprint

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io/fs"

	"github.com/twmb/murmur3"

	"github.com/sargassum-eco/fluitans/pkg/godest/fsutil"
)

const FingerprintSize = 8

func Compute(data []byte) string {
	hash := murmur3.Sum64(data)
	hashBytes := make([]byte, FingerprintSize)
	binary.LittleEndian.PutUint64(hashBytes, hash)
	return fmt.Sprintf("%x:%s", len(data), base64.StdEncoding.EncodeToString(hashBytes))
}

func ComputeFiles(filenames []string, f fs.FS) (map[string]string, error) {
	fingerprints := make(map[string]string)
	for _, name := range filenames {
		data, err := fsutil.ReadFile(name, f)
		if err != nil {
			return nil, err
		}
		fingerprints[name] = Compute(data)
	}
	return fingerprints, nil
}

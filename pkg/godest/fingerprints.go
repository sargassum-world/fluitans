package godest

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"

	"github.com/twmb/murmur3"
)

const fingerprintSize = 8

func computeFingerprint(data []byte) string {
	hash := murmur3.Sum64(data)
	hashBytes := make([]byte, fingerprintSize)
	binary.LittleEndian.PutUint64(hashBytes, hash)
	return fmt.Sprintf("%x:%s", len(data), base64.StdEncoding.EncodeToString(hashBytes))
}

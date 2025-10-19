package scylla_migrator

import (
	"crypto/md5"
	"encoding/hex"
)

var encode = hex.EncodeToString

func checksum(b []byte) string {
	v := md5.Sum(b)
	return encode(v[:])
}

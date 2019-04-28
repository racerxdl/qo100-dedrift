package rpc

import (
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/sha3"
)

func GeneratePassword(salt, password, timestamp string) string {
	s := sha3.New512()
	h := fmt.Sprintf("%s%s%s", timestamp, salt, password)
	_, _ = s.Write([]byte(h))

	return hex.EncodeToString(s.Sum(nil))
}

func ComparePassword(salt, hash, timestamp, expectedPassword string) bool {
	generated := GeneratePassword(salt, timestamp, expectedPassword)
	return generated == hash
}

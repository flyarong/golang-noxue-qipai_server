package utils

import (
	"crypto/md5"
	"fmt"
)

func PassEncode(password string) string {
	has := md5.Sum([]byte(password))
	return fmt.Sprintf("%x", has)
}

func PassCompare(password, hashStr string) bool {
	if PassEncode(password) == hashStr {
		return true
	}
	return false
}

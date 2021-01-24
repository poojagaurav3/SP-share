package common

import (
	"crypto/sha256"
	"fmt"
	"strconv"
)

// ExractUserID is a helper method to convert user id from string to integer
func ExractUserID(userID string) int64 {
	if userID == "" {
		return -1
	}

	intUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return -1
	}

	return intUserID
}

// SHA256 returns the sha 256 hash of the string
func SHA256(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil))
}

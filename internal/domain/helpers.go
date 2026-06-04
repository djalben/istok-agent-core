package domain

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"
)

// GenerateID генерирует уникальный ID.
func GenerateID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	return hex.EncodeToString(bytes)
}

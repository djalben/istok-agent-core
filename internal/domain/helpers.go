package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateID генерирует уникальный ID
func GenerateID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

package utils

import (
	"github.com/google/uuid"
)

// GenerateUUID 生成標準的 UUID v4
func GenerateUUID() string {
	return uuid.New().String()
}

// IsValidUUID 驗證 UUID 格式是否正確
func IsValidUUID(uuidStr string) bool {
	_, err := uuid.Parse(uuidStr)
	return err == nil
}

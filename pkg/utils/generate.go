package utils

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// ==================== UUID & TOKEN ====================

func GenerateUUID() uuid.UUID {
	return uuid.New()
}

func GenerateUUIDString() string {
	return uuid.New().String()
}

func ParseUUID(uuidStr string) (uuid.UUID, error) {
	return uuid.Parse(uuidStr)
}

func GenerateSessionToken() uuid.UUID {
	return uuid.New()
}

// ==================== OTP ====================

func GenerateOTP(length int) string {
	if length <= 0 {
		length = 6
	}

	rand.New(rand.NewSource(time.Now().UnixNano()))

	otp := ""
	for i := 0; i < length; i++ {
		otp += fmt.Sprintf("%d", rand.Intn(10))
	}

	return otp
}

// ==================== ORDER ID ====================

func GenerateOrderID() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	now := time.Now()

	// Format: BOOK-YYYYMMDD-HHMMSS-RANDOM
	datePart := now.Format("20060102")
	timePart := now.Format("150405")
	randomPart := fmt.Sprintf("%04d", rand.Intn(10000))

	return fmt.Sprintf("BOOK-%s-%s-%s", datePart, timePart, randomPart)
}

package utils

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// ParseInt converts string to int with default value
func ParseInt(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	if result < 1 {
		return defaultValue
	}

	return result
}

// GenerateOTP creates a numeric OTP of specified length
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

// GenerateOrderID creates a unique order ID with timestamp
func GenerateOrderID() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	now := time.Now()

	// Format: BOOK-YYYYMMDD-HHMMSS-RANDOM
	datePart := now.Format("20060102")
	timePart := now.Format("150405")
	randomPart := fmt.Sprintf("%04d", rand.Intn(10000))

	return fmt.Sprintf("BOOK-%s-%s-%s", datePart, timePart, randomPart)
}

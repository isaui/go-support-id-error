package errorid

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateErrorID creates a unique error ID with format: ERR-YYYYMMDD-XXXXXX
// XXXXXX is a random hex string for collision resistance
func GenerateErrorID() string {
	date := time.Now().Format("20060102")
	randomBytes := make([]byte, 3) // 3 bytes = 6 hex chars
	
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Fallback to timestamp-based if crypto/rand fails
		return fmt.Sprintf("ERR-%s-%06d", date, time.Now().UnixNano()%1000000)
	}
	
	randomHex := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("ERR-%s-%s", date, randomHex)
}

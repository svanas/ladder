package oneinch

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

func stringToHexBytes(s string) ([]byte, error) {
	// strip the "0x" prefix if it exists
	cleaned := trimPrefix(s, "0x")

	// ensure the string has an even length by padding with a zero if it's odd
	if len(cleaned)%2 != 0 {
		cleaned = "0" + cleaned
	}

	// decode the string into bytes
	bytes, err := hex.DecodeString(cleaned)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func generateSalt(extension string, useRandom bool) (*big.Int, error) {
	salt := big.NewInt(0)

	// generate upper 32 bits (bits 224-255) - tracking code mask
	trackingHash := crypto.Keccak256Hash([]byte("sdk"))
	trackingCodeMask := new(big.Int).Lsh(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 32), big.NewInt(1)), 224) // (2^32 - 1) << 224
	trackingBits := new(big.Int).SetBytes(trackingHash.Bytes())
	trackingBits.And(trackingBits, trackingCodeMask)
	salt.Or(salt, trackingBits)

	// generate middle 64 bits (bits 160-223)
	var middleBits *big.Int
	if useRandom {
		// Generate random 64 bits
		randomBytes := make([]byte, 8)
		_, err := rand.Read(randomBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random bytes: %w", err)
		}
		middleBits = new(big.Int).SetBytes(randomBytes)
	} else {
		middleBits = big.NewInt(time.Now().Unix())
	}
	// mask to 64 bits and shift to position 160-223
	mask64 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 64), big.NewInt(1))
	middleBits.And(middleBits, mask64)
	middleBits.Lsh(middleBits, 160)
	salt.Or(salt, middleBits)

	// handle extension for lower 160 bits
	if extension == "0x" || extension == "" {
		// if there is no extension, salt can be anything for the lower bits
		// (middle bits already set above, lower 160 bits remain 0 or can be left as-is)
	} else {
		// lower 160 bits must be from keccak256 hash of the extension
		extensionBytes, err := stringToHexBytes(extension)
		if err != nil {
			return nil, fmt.Errorf("failed to convert extension to bytes: %w", err)
		}
		extensionHash := crypto.Keccak256Hash(extensionBytes)
		mask160 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 160), big.NewInt(1))
		extensionBits := new(big.Int).SetBytes(extensionHash.Bytes())
		extensionBits.And(extensionBits, mask160)
		salt.Or(salt, extensionBits)
	}

	return salt, nil
}

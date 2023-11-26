package uuid

import (
	"crypto/rand"
	"fmt"
)

type UUID struct {
	inner []byte
}

func New() *UUID {
	uuid := UUID{inner: make([]byte, 16)}
	_, err := rand.Read(uuid.inner)
	if err != nil {
		panic(fmt.Errorf("error generating uuid: %v", err))
	}
	// variant bits; see section 4.1.1
	uuid.inner[8] = uuid.inner[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid.inner[6] = uuid.inner[6]&^0xf0 | 0x40
	return &uuid
}

func (uuid *UUID) String() string {
	return fmt.Sprintf("%X-%X-%X-%X-%X", uuid.inner[0:4], uuid.inner[4:6], uuid.inner[6:8], uuid.inner[8:10], uuid.inner[10:])
}

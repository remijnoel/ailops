package internal

import (
	"github.com/google/uuid"
)

func GenerateUniqueID() string {
	// Returns a uuid
	newUUID, err := uuid.NewRandom()
	if err != nil {
		panic("Failed to generate unique ID: " + err.Error())
	}
	return newUUID.String()
}

package domain

type Error struct {

	// The error message
	// Required: true
	Message *string `json:"message"`
}

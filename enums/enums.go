package enums

// ResponseStatus enum
type ResponseStatus int

const (
	// Success response
	Success ResponseStatus = iota
	// Error response
	Error
)

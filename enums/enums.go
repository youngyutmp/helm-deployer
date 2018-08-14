package enums

// ResponseStatus enum
type ResponseStatus int

const (
	// StatusSuccess defines successful response
	StatusSuccess ResponseStatus = iota
	// StatusError defines failed response
	StatusError
)

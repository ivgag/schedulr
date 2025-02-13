package domain

type error interface {
	Error() string
}

type NotFoundError struct {
	Message string
}

// Error implements error.
func (e NotFoundError) Error() string {
	return e.Message
}

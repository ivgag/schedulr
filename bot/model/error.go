package model

type Error interface {
	Error() string
}

type errorImpl struct {
	message string
}

func (e errorImpl) Error() string {
	return e.message
}

func ErrorForMessage(message string) Error {
	return errorImpl{message: message}
}

type NotFoundError struct {
	Message string
}

// Error implements error.
func (e NotFoundError) Error() string {
	return e.Message
}

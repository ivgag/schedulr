package ai

type ApiError struct {
	Message      string
	ResponseCode int
	Retryable    bool
}

func (e ApiError) Error() string {
	return e.Message
}

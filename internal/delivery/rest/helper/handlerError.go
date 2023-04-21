package helper

// HandlerError is an error with a message and a response code that can be sent to the user.
type HandlerError struct {
	Message string
	Code    int
}

func (e *HandlerError) Error() string {
	return e.Message
}

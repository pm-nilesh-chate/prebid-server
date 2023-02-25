package errorcodes

// ErrorCode new type defined for wrapper errors
type ErrorCode = int

// IError Interface for Custom Errors
type IError interface {
	error
	Code() ErrorCode
}

// Error Structure for Custom Errors
type Error struct {
	IError
	code    ErrorCode
	message string
}

// NewError New Object
func NewError(code ErrorCode, message string) IError {
	return &Error{code: code, message: message}
}

// Code Returns Error Code
func (e *Error) Code() ErrorCode {
	return e.code
}

// Error Returns Error Message
func (e *Error) Error() string {
	return e.message
}

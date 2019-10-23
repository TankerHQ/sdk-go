package core

// ErrorCode represents a Tanker error code.
type ErrorCode uint32

const (
	ErrorInvalidArgument ErrorCode = iota + 1
	ErrorInternalError
	ErrorNetworkError
	ErrorPreconditionFailed
	ErrorOperationCanceled

	ErrorDecryptionFailed

	ErrorGroupTooBig

	ErrorInvalidVerification
	ErrorTooManyAttempts
	ErrorExpiredVerification
	ErrorIoError
)

// Error is the Tanker error interface. Cast the error returned by
// Tanker functions to this interface to get more informations.
type Error interface {
	error
	Code() ErrorCode
}

type tankerError struct {
	code    ErrorCode
	message string
}

func newError(code ErrorCode, message string) error {
	return &tankerError{code, message}
}

func (e tankerError) Error() string {
	return e.message
}

func (e tankerError) Code() ErrorCode {
	return e.code
}

package errors

// Universal error type used to distinct all errors returned by services
type ServiceError struct {
	// error message
	message string

	// ineer error if any
	innerError error

	// error kind. Used to distinct between reasons error occured
	kind int
}

// Retuns string message of current error
func (err ServiceError) Error() string {
	return err.message
}

// Returns current error kind
func (err ServiceError) Kind() int {
	return err.kind
}

// Returns inner error if any
func (err ServiceError) Unwrap() error {
	return err.innerError
}

// Checks if current error of the same type as target error
//	target - target error
func (err ServiceError) Is(target error) bool {
	if err == target {
		return true
	}

	if target == nil {
		return false
	}

	var typedErr, ok = target.(ServiceError)
	return ok && typedErr.kind == err.kind
}

// Error kind - DB error. Used to wrap around errors, returned by DB driver
const ErrorKindDB int = 1

// Returns new service error
//	message    - error message
//	innerError - inner error, if any
//	kind       - error kind
// Returns created service error
func NewServiceError(message string, innerError error, kind int) error {
	return ServiceError{message, innerError, kind}
}

// Returns new service error wrapping around database error
func ErrDatabaseError(dbError error) error {
	return NewServiceError("error occured when trying to work with database", dbError, ErrorKindDB)
}

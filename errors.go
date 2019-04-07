package docshelf

// ErrDoesNotExist is a special error type for signaling that a requested entity
// doesn't exist.
type ErrDoesNotExist struct {
	msg string
}

// ErrRemoved is a special error type for signaling that a requested entity
// has been removed.
type ErrRemoved struct {
	msg string
}

// NewErrDoesNotExist returns a new ErrDoesNotExist as a normal error containing
// the given message.
func NewErrDoesNotExist(msg string) error {
	return ErrDoesNotExist{msg}
}

// NewErrRemoved returns a new ErrRemoved as a normal error containing the given
// message.
func NewErrRemoved(msg string) error {
	return ErrRemoved{msg}
}

// Error implements the Error interface for ErrDoesNotExist. Default messaging is
// used if not supplied.
func (e ErrDoesNotExist) Error() string {
	if e.msg == "" {
		return "entity does not exist"
	}

	return e.msg
}

// Error implements the Error interface for ErrRemoved. Default messaging is
// used if not supplied.
func (e ErrRemoved) Error() string {
	if e.msg == "" {
		return "entity is removed"
	}

	return e.msg
}

// CheckDoesNotExist is a helper function for determining if an error type is
// actually an ErrDoesNotExist.
func CheckDoesNotExist(err error) bool {
	_, ok := err.(ErrDoesNotExist)
	return ok
}

// CheckRemoved is a helper function for determining if an error type is
// actually an ErrRemoved.
func CheckRemoved(err error) bool {
	_, ok := err.(ErrRemoved)
	return ok
}

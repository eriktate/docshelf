package docshelf

// ErrNotFound is a special error type for signaling that a requested entity
// doesn't exist.
type ErrNotFound struct {
	msg string
}

// ErrRemoved is a special error type for signaling that a requested entity
// has been removed.
type ErrRemoved struct {
	msg string
}

// NewErrNotFound returns a new ErrNotFound as a normal error containing
// the given message.
func NewErrNotFound(msg string) error {
	return ErrNotFound{msg}
}

// NewErrRemoved returns a new ErrRemoved as a normal error containing the given
// message.
func NewErrRemoved(msg string) error {
	return ErrRemoved{msg}
}

// Error implements the Error interface for ErrNotFound. Default messaging is
// used if not supplied.
func (e ErrNotFound) Error() string {
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

// CheckNotFound is a helper function for determining if an error type is
// actually an ErrNotFound.
func CheckNotFound(err error) bool {
	_, ok := err.(ErrNotFound)
	return ok
}

// CheckRemoved is a helper function for determining if an error type is
// actually an ErrRemoved.
func CheckRemoved(err error) bool {
	_, ok := err.(ErrRemoved)
	return ok
}

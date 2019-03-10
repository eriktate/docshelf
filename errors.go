package skribe

type ErrDoesNotExist struct {
	msg string
}

type ErrRemoved struct {
	msg string
}

func NewErrDoesNotExist(msg string) error {
	return ErrDoesNotExist{msg}
}

func NewErrRemoved(msg string) error {
	return ErrRemoved{msg}
}

func (e ErrDoesNotExist) Error() string {
	if e.msg == "" {
		return "entity does not exist"
	}

	return e.msg
}

func (e ErrRemoved) Error() string {
	if e.msg == "" {
		return "entity is removed"
	}

	return e.msg
}

func CheckDoesNotExist(err error) bool {
	_, ok := err.(ErrDoesNotExist)
	return ok
}

func CheckRemoved(err error) bool {
	_, ok := err.(ErrRemoved)
	return ok
}

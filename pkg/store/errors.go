package store

type Error string

func (e Error) Error() string {
	return string(e)
}

const (
	ErrAlreadyExists     = Error("already exists")
	ErrNotFound          = Error("not found")
	ErrInvalidDependency = Error("record dependency is invalid")
	ErrUnknown           = Error("unknown error")
)

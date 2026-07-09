package todo

import "github.com/oklog/ulid/v2"

type ID ulid.ULID

func NewID() ID {
	id := ulid.Make()
	return ID(id)
}

func ParseID(s string) (ID, error) {
	id, err := ulid.ParseStrict(s)
	if err != nil {
		return ID{}, ErrInvalid
	}
	return ID(id), nil
}

func (id ID) String() string {
	return ulid.ULID(id).String()
}

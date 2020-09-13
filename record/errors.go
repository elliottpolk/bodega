package record

import "errors"

var (
	ErrImmutableValue          = errors.New("attempting to edit immutable value")
	ErrInvalidRecordInfo       = errors.New("invalid record info")
	ErrInvalidCreatedBy        = errors.New("invalid created by value")
	ErrInvalidUpdatedBy        = errors.New("invalid updated by value")
	ErrDuplicateRecord         = errors.New("duplicate record for provided id")
	ErrNotFound                = errors.New("no valid record for provided id")
	ErrInvalidId               = errors.New("no valid id for provided record")
	ErrInvalidUsername         = errors.New("no valid username provided")
	ErrIncompleteAction        = errors.New("not all records were properly acted on")
	ErrMutlipleRecordsReturned = errors.New("multiple records returned")
	ErrNoMongoClient           = errors.New("no valid mongo client")
)

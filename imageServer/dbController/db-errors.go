package dbController

// Used for when there's an issue with a database
type DBError struct{ ErrMsg string }

func (err DBError) Error() string { return err.ErrMsg }
func NewDBError(msg string) error { return DBError{msg} }

type NoResultsError struct{ ErrMsg string }

func (err NoResultsError) Error() string { return err.ErrMsg }
func NewNoResultsError(msg string) error { return NoResultsError{msg} }

type DuplicateEntryError struct{ ErrMsg string }

func (err DuplicateEntryError) Error() string { return err.ErrMsg }
func NewDuplicateEntryError(msg string) error { return DuplicateEntryError{msg} }

type InvalidInputError struct{ ErrMsg string }

func (err InvalidInputError) Error() string { return err.ErrMsg }
func NewInvalidInputError(msg string) error { return InvalidInputError{msg} }

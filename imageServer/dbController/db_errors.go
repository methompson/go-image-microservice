package dbController

// Used for when there's an issue with a database
type DBError struct{ ErrMsg string }

func (err DBError) Error() string { return err.ErrMsg }
func NewDBError(msg string) error { return DBError{msg} }

// Used to communicate that no results exist for a given input
type NoResultsError struct{ ErrMsg string }

func (err NoResultsError) Error() string { return err.ErrMsg }
func NewNoResultsError(msg string) error { return NoResultsError{msg} }

// Used to communicate that this value cannot be inserted, becuase
// another entry exists with the specific index.
type DuplicateEntryError struct{ ErrMsg string }

func (err DuplicateEntryError) Error() string { return err.ErrMsg }
func NewDuplicateEntryError(msg string) error { return DuplicateEntryError{msg} }

// Used to communicate that the value provided as an index cannot be parsed
type InvalidInputError struct{ ErrMsg string }

func (err InvalidInputError) Error() string { return err.ErrMsg }
func NewInvalidInputError(msg string) error { return InvalidInputError{msg} }

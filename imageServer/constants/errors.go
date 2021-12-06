package constants

type NoTokenError struct{ ErrMsg string }

func (err NoTokenError) Error() string { return err.ErrMsg }
func NewNoTokenError(msg string) error { return NoTokenError{msg} }

type InvalidTokenError struct{ ErrMsg string }

func (err InvalidTokenError) Error() string { return err.ErrMsg }
func NewInvalidTokenError(msg string) error { return InvalidTokenError{msg} }

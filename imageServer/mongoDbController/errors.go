package mongoDbController

// Used for when there's an issue with reading Nonces
type EnvironmentVariableError struct{ ErrMsg string }

func (err EnvironmentVariableError) Error() string { return err.ErrMsg }
func NewEnvironmentVariableError(msg string) error { return EnvironmentVariableError{msg} }

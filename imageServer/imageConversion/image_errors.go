package imageConversion

type ImageError struct{ ErrMsg string }

func (err ImageError) Error() string { return err.ErrMsg }
func NewDBError(msg string) error    { return ImageError{msg} }

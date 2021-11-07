package user

type UserType int

const (
	Reader UserType = iota
	Editor
	Admin
)

func (s UserType) String() string {
	switch s {
	case Reader:
		return "reader"
	case Editor:
		return "editor"
	case Admin:
		return "admin"
	}

	return "unknown"
}

type UserInformation struct {
	Uid    string
	Name   string
	Email  string
	Active bool
	Role   UserType
}

func (ui UserInformation) Json() map[string]interface{} {
	return map[string]interface{}{
		"uid":    ui.Uid,
		"name":   ui.Name,
		"email":  ui.Email,
		"active": ui.Active,
		"role":   ui.Role.String(),
	}
}

package auth

type ErrUserAlreadyRegistered struct {
}

func (e ErrUserAlreadyRegistered) Error() string {
	return "user already registered"
}


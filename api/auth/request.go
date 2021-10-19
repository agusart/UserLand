package auth

import "userland/api"

type RegisterRequest struct {
	FullName string `json:"full_name"`
	Email string `json:"email"`
	Password string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

type VerificationRequest struct {
	Type      string `json:"type"`
	Recipient string `json:"recipient"`
}

func (r RegisterRequest) Validate() map[string]string {
	errorMsg := make(map[string]string)

	if len(r.FullName) < 5 || len(r.FullName) > 40 {
		errorMsg["full_name"] = "Fullname must between 5 to 40 character"
	}

	if !isValidEmail(r.Email) {
		errorMsg["email"] = "Email is invalid"
	}

	if r.PasswordConfirm != r.Password {
		errorMsg["password_confirm"] = "Password confirm is invalid"
		return errorMsg
	}

	if !isValidPassword(r.Password) {
		errorMsg["password"] = "password is not strong"
	}

	return errorMsg
}


func (r VerificationRequest) Validate() map[string]string {
	errorMsg := make(map[string]string)

	if !isValidEmail(r.Recipient) {
		errorMsg["email"] = "Email is invalid"
	}

	if r.Type != api.ActionVerifyEmail {
		errorMsg["email"] = "invalid type"
	}

	return errorMsg
}
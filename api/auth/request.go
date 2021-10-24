package auth

import "userland/api"

type RegisterRequest struct {
	FullName string `json:"full_name"`
	Email string `json:"email"`
	Password string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
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

type RegisterVerificationRequest struct {
	Type      string `json:"type"`
	Recipient string `json:"recipient"`
}

func (r RegisterVerificationRequest) Validate() map[string]string {
	errorMsg := make(map[string]string)

	if !isValidEmail(r.Recipient) {
		errorMsg["recipient"] = "recipient is invalid"
	}

	if r.Type != api.ActionVerifyEmail {
		errorMsg["type"] = "invalid type"
	}

	return errorMsg
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

func (f ForgotPasswordRequest) Validate() map[string]string {
	errorMsg := make(map[string]string)

	if !isValidEmail(f.Email) {
		errorMsg["email"] = "invalid email format"
	}

	return errorMsg
}

type ResetPasswordRequest struct {
	 Token string `json:"token"`
	 Password string `json:"password"`
	 PasswordConfirm string `json:"password_confirm"`
}

func (r ResetPasswordRequest) Validate() map[string]string {
	errorMsg := make(map[string]string)

	if r.Token == "" {
		errorMsg["token"] = "token should not be empty"
	}

	if !isValidPassword(r.Password) {
		errorMsg["password"] = "password is weak"
	}

	if r.Password != r.PasswordConfirm {
		errorMsg["password_confirm"] = "password confirm not match"
	}

	return errorMsg
}

type LoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}


func(r LoginRequest) Validate() map[string]string{
	errorMsg := make(map[string]string)

	if !isValidEmail(r.Email) {
		errorMsg["email"] = "email is invalid"
	}

	if len(r.Password) < 1 {
		errorMsg["password"] = "password must not be empty"
	}

	return errorMsg

}

type VerifyTfaRequest struct {
	Code string `json:"code"`
}

func(r VerifyTfaRequest) Validate() map[string]string {
	errorMsg := make(map[string]string)

	if r.Code == "" {
		errorMsg["code"] = "code required"
	}

	return errorMsg
}
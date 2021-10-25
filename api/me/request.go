package me

import (
	"fmt"
	"net/url"
	"userland/api/auth"
	"userland/store/postgres"
)

type UserUpdateRequest interface {
	Validate() map[string]string
	UpdateUser(existedUser postgres.User) postgres.User
}

type UpdateUserDetailRequest struct {
	FullName string `json:"fullname"`
	Location string `json:"location"`
	Bio string `json:"bio"`
	Web string `json:"web"`
}

func (u UpdateUserDetailRequest) UpdateUser(existedUser postgres.User) postgres.User {
	existedUser.Web = u.Web
	existedUser.Bio = u.Bio
	existedUser.Location = u.Location
	existedUser.FullName = u.FullName

	return existedUser
}

func (u UpdateUserDetailRequest) Validate() map[string]string {
	errMsg := make(map[string]string)

	if len(u.FullName) < 2 &&  len(u.FullName) > 50 {
		errMsg["fullname"]  = "fullname only between 2 - 50 char"
	}

	if len(u.Location) > 100 {
		errMsg["location"]  = "location no more than 100 char"
	}

	if len(u.Bio) > 200 {
		errMsg["location"]  = "location no more than 200 char"
	}

	_, err := url.ParseRequestURI(u.Web)
	if err!= nil {
		errMsg["web"] = fmt.Sprintf("%s is not valid website", u.Web)
	}

	return errMsg
}


type EmailAddressChangeRequset struct {
	Email string `json:"email"`
}

func (e EmailAddressChangeRequset) Validate() map[string]string{
	errMsg := make(map[string]string)

	if !auth.IsValidEmail(e.Email){
		errMsg["email"] = "email submitted is onvalid"
	}

	return errMsg
}

func (e EmailAddressChangeRequset) UpdateUser(existedUser postgres.User) postgres.User {
	existedUser.Email = e.Email

	return existedUser
}


type UpdatePasswordRequest struct {
	PasswordCurrent string `json:"password_current"`
	PasswordNew string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

func (u UpdatePasswordRequest) Validate() map[string]string {
	errMsg := make(map[string]string)

	if !auth.IsValidPassword(u.PasswordCurrent){
		errMsg["password_current"] = "current password is invalid"
	}

	if u.PasswordConfirm != u.PasswordNew {
		errMsg["password_confirm"] = "confirmation password not equal new password"
		return errMsg
	}

	if !auth.IsValidPassword(u.PasswordNew){
		errMsg["password"] = "Invalid new password "
	}
	return errMsg
}

func (u UpdatePasswordRequest) UpdateUser(existedUser postgres.User) postgres.User {
	password, err := auth.HashPassword(u.PasswordNew)
	if err != nil {
		return existedUser
	}

	existedUser.Password = password
	return existedUser
}


type ActivateTfaRequest struct {
	Secret string `json:"secret"`
	Code string `json:"code"`
}

func (a ActivateTfaRequest) Validate() map[string]string {
	errMsg := make(map[string]string)

	if a.Secret == "" {
		errMsg["secret"] = "secret tfa is nil"
	}

	if len(a.Code) != 8 {
		errMsg["code"] = "code is not 8 digit"
	}

	return errMsg
}


type RemoveTfaRequest struct {
	 Password string `json:"password"`
}

func (u RemoveTfaRequest) Validate() map[string]string {
	errMsg := make(map[string]string)

	if !auth.IsValidPassword(u.Password){
		errMsg["password"] = "password format password is invalid"
	}

	return errMsg
}


type DeleteAccountRequest struct {
	Password string `json:"password"`
}

func (d DeleteAccountRequest) Validate() map[string]string {
	errMsg := make(map[string]string)

	if !auth.IsValidPassword(d.Password){
		errMsg["password"] = "password format password is invalid"
	}

	return errMsg
}

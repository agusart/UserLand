package me

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"userland/api"
	"userland/api/auth"
	"userland/api/middleware"
	"userland/store/postgres"
)

func UserDetail(userStore postgres.UserStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claim := r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)
		user, err := userStore.GetUserById(claim.UserId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(NewUserResponse(*user))
		return
	}
}


func UpdateUserDetail(userStore postgres.UserStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := UpdateUserDetailRequest{}
		updatedUser, err := getUpdatedUserFromRequest(request, userStore, w, r)
		if err != nil {
			return
		}
		err = userStore.UpdateUserBasicInfo(*updatedUser)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"status" : true,
		})

	}
}


func GetCurrentEmailAddress(userStore postgres.UserStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claim := r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)
		existedUser, err := userStore.GetUserById(claim.UserId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		if existedUser == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"email" : existedUser.Email,
		})
	}
}


func UpdateUserEmail(userStore postgres.UserStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := EmailAddressChangeRequset{}
		updatedUser, err := getUpdatedUserFromRequest(request, userStore, w, r)
		if err != nil {
			return
		}

		err = userStore.UpdateUserBasicInfo(*updatedUser)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"status" : true,
		})

	}
}

func UpdateUserPassword(userStore postgres.UserStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := UpdatePasswordRequest{}
		updatedUser, err := getUpdatedUserFromRequest(request, userStore, w, r)
		if err != nil {
			return
		}

		err = userStore.UpdateUserPassword(updatedUser.Id, updatedUser.Password)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"status" : true,
		})
	}
}

func SetupTfa(userStore postgres.UserStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claim := r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)
		user, err := userStore.GetUserById(claim.UserId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		if user == nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		secret := generateTfaSecret()
		tfaLink := generateAuthLink(secret, user.Email)
		qrString, err := GenerateQRString(tfaLink)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"secret" : secret,
			"qr" : qrString,
		})
	}
}

func ActivateTfa(
	userStore postgres.UserStoreInterface,
	) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claim := r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)
		request := ActivateTfaRequest{}
		_ = json.NewDecoder(r.Body).Decode(&request)

		if errMsg := request.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(api.Response{
				"fields": errMsg,
			})
		}

		success, err  := verifyTfaCode(request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = userStore.SaveUserTfaSecret(request.Secret, claim.UserId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		_ = json.NewEncoder(w).Encode(api.Response{
			"success": success,
		})
	}
}


func RemoveTfa(
	userStore postgres.UserStoreInterface,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claim := r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)
		request := RemoveTfaRequest{}
		_ = json.NewDecoder(r.Body).Decode(&request)

		if errMsg := request.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(api.Response{
				"fields": errMsg,
			})
		}

		logedInUser, err := userStore.GetUserById(claim.UserId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		if logedInUser == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !auth.CheckPasswordHash(request.Password, logedInUser.Password) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(postgres.CustomError{
				StatusCode: api.ErrWrongPasswordCode,
				Err: errors.New("invalid password"),
			})
			return
		}

		err = userStore.RemoveTfaStatus(claim.UserId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		_ = json.NewEncoder(w).Encode(api.Response{
			"success": true,
		})
	}
}

func DeleteAccount(userStore postgres.UserStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claim := r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)
		request := DeleteAccountRequest{}
		_ = json.NewDecoder(r.Body).Decode(&request)

		if errMsg := request.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(api.Response{
				"fields": errMsg,
			})
		}

		logedInUser, err := userStore.GetUserById(claim.UserId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		if logedInUser == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !auth.CheckPasswordHash(request.Password, logedInUser.Password) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(postgres.CustomError{
				StatusCode: api.ErrWrongPasswordCode,
				Err: errors.New("invalid password"),
			})
			return
		}

		err = userStore.DeleteUser(logedInUser.Id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"success" : true,
		})
	}
}

func getUpdatedUserFromRequest(
		request UserUpdateRequest,
		userStore postgres.UserStoreInterface,
		w http.ResponseWriter,
		r *http.Request,
	) (*postgres.User, error){
		claim := r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)
		_ = json.NewDecoder(r.Body).Decode(&request)

		if errMsg := request.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(api.Response{
				"fields": errMsg,
			})
			return nil, fmt.Errorf("request has invalid input")
		}

		existedUser, err := userStore.GetUserById(claim.UserId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return nil, err
		}

		if existedUser == nil {
			w.WriteHeader(http.StatusBadRequest)
			return nil, fmt.Errorf("existed user is nil")
		}

		updatedUser := request.UpdateUser(*existedUser)

		return &updatedUser, nil
}



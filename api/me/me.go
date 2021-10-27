package me

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"userland/api"
	"userland/api/auth"
	"userland/api/middleware"
	"userland/store/postgres"
	"userland/store/redis"
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
		claim := r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)

		_ = json.NewDecoder(r.Body).Decode(&request)
		if errMsg := request.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(api.Response{
				"fields": errMsg,
			})

			return
		}

		updatedUser, err := getUpdatedUserFromRequest(request, userStore, claim)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		if updatedUser == nil {
			w.WriteHeader(http.StatusBadRequest)
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
			"status": true,
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
			"email": existedUser.Email,
		})
	}
}

func UpdateUserEmailRequest(cache redis.CacheInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := EmailAddressChangeRequset{}
		claim := r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)

		_ = json.NewDecoder(r.Body).Decode(&request)
		if errMsg := request.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(api.Response{
				"fields": errMsg,
			})
			return
		}

		err := cache.RequestChangeEmail(r.Context(), claim.UserId, request.Email, api.VerificationExpiredTime)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"status": true,
		})

	}
}


func UpdateUserEmail(userStore postgres.UserStoreInterface, cache redis.CacheInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		verifyToken := chi.URLParam(r, "verifyToken")
		if verifyToken == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		claim := r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)
		email, err := cache.GetVerifyChangeEmail(r.Context(), claim.UserId, strings.TrimSpace(verifyToken))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		err = userStore.ChangeUserEmail(claim.UserId, email)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}


		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"status": true,
		})

	}
}

func UpdateUserPassword(userStore postgres.UserStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := UpdatePasswordRequest{}
		claim := r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)

		_ = json.NewDecoder(r.Body).Decode(&request)
		if errMsg := request.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(api.Response{
				"fields": errMsg,
			})

			return
		}

		updatedUser, err := getUpdatedUserFromRequest(request, userStore, claim)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))

			return
		}

		oldPasswordList, err := userStore.GetPasswordHistory(claim.UserId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))

			return
		}

		olPasswordDetected := false
		for _, p := range oldPasswordList {
			if auth.CheckPasswordHash(request.PasswordNew, p) {
				olPasswordDetected = true
				break
			}
		}

		if olPasswordDetected {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.Response{
				"messages" : "use password differently from your 3 last password",
			})
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
			"status": true,
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
			"secret": secret,
			"qr":     qrString,
		})
	}
}

func ActivateTfa(
	tfaStore postgres.TfaStoreInterface,
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

			return
		}

		success, err := api.VerifyTfaCode(request.Secret, request.Code)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = tfaStore.SaveUserTfaSecret(request.Secret, claim.UserId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}
		_, err = tfaStore.CreateTfaBackupCode(claim.UserId)
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
	tfaStore postgres.TfaStoreInterface,
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
				Err:        errors.New("invalid password"),
			})
			return
		}

		err = tfaStore.RemoveTfaStatus(claim.UserId)
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

func GetCurrentTfaStatus(userStore postgres.UserStoreInterface) http.HandlerFunc {
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
			"enabled": existedUser.TfaEnabled,
			"enabled_at" : "",
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
				Err:        errors.New("invalid password"),
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
			"success": true,
		})
	}
}

func getUpdatedUserFromRequest(
	request UserUpdateRequest,
	userStore postgres.UserStoreInterface,
	claim middleware.JWTClaims,
) (*postgres.User, error) {
	existedUser, err := userStore.GetUserById(claim.UserId)
	if err != nil {
		return nil, err
	}

	if existedUser == nil {
		return nil, fmt.Errorf("existed user is nil")
	}

	updatedUser := request.UpdateUser(*existedUser)
	return &updatedUser, nil
}

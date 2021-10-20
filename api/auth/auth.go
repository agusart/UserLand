package auth

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"userland/api"
	"userland/store/postgres"
)

func Login(userStore postgres.UserStoreInterface, authStore postgres.AuthStoreInterface) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

	}
}

func Register(userStore postgres.UserStoreInterface, authStore postgres.AuthStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registerRequest := RegisterRequest{}
		_ = json.NewDecoder(r.Body).Decode(&registerRequest)

		if errMsg := registerRequest.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(Response{
				"fields": errMsg,
			})
			return
		}

		user := NewUserFromRegisterRequest(registerRequest)
		err := userStore.RegisterUser(user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			errResponse := GenerateErrorResponse(err)
			_ = json.NewEncoder(w).Encode(errResponse)
			return
		}

		token, err := authStore.SendRegistrationVerificationCode(r.Context(), user.Email, api.VerificationExpiredTime)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		//go func() {
		//	api.SendEmail(verificationRequest.Recipient, token)
		//}()

		log.Printf("token : %v", token)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Response{
			"success": true,
		})
	}
}

func RequestVerification(userStore postgres.UserStoreInterface, authStore postgres.AuthStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		verificationRequest := RegisterVerificationRequest{}
		_ = json.NewDecoder(r.Body).Decode(&verificationRequest)

		if errMsg := verificationRequest.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(Response{
				"fields": errMsg,
			})
			return
		}

		isVerified, err := userStore.IsUserVerified(verificationRequest.Recipient)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		if isVerified {
			_ = json.NewEncoder(w).Encode(
				ErrorResponse{
					Code: postgres.ErrUserAlreadyVerifiedCode,
					Message: "User Already Verified",
				},
			)
			return
		}
		//
		token, err := authStore.SendRegistrationVerificationCode(r.Context(), verificationRequest.Recipient, api.VerificationExpiredTime)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		//go func() {
		//	api.SendEmail(verificationRequest.Recipient, token)
		//}()

		log.Printf("token : %v", token)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Response{
			"success": true,
		})
	}
}

func VerifyRegister(userStore postgres.UserStoreInterface, authStore postgres.AuthStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		if token == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(Response{
				"code" : "ERR-11",
				"messages" : "error link",
			})

			return
		}

		email, err := authStore.GetRegistrationCodeEmail(r.Context(), token)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(Response{
				"code" : "ERR-",
				"messages" : "error link",
			})
			return
		}

		err = userStore.VerifyUser(email)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(Response{
				"code" : "ERR-",
				"messages" : "error link",
			})

			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Response{
			"success": true,
		})
	}
}

func ForgetPassword(userStore postgres.UserStoreInterface, authStore postgres.AuthStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		forgetPasswordRequest := ForgotPasswordRequest{}
		json.NewEncoder(w).Encode(&forgetPasswordRequest)

		if errMsg := forgetPasswordRequest.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(Response{
				"fields": errMsg,
			})
			return
		}

		existingUser, err := userStore.GetUserByEmail(forgetPasswordRequest.Email)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		if existingUser == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if existingUser.DeletedAt != nil && !existingUser.Verified {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				api.ErrBadRequestErrorCode,
				"cant reset password",
			})
			return
		}

		token, err := authStore.SendForgotPasswordVerificationCode(r.Context(), existingUser.Email, api.ForgotPasswordExpiredTime)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		//go func() {
		//	api.SendEmail(forgetPasswordRequest.Email, token)
		//}()

		log.Printf("token : %v", token)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Response{
			"success": true,
		})

	}
}

func ResetPassword(userStore postgres.UserStoreInterface, authStore postgres.AuthStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := ResetPasswordRequest{}
		json.NewEncoder(w).Encode(&request)

		if errMsg := request.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(Response{
				"fields": errMsg,
			})
			return
		}

		email, err := authStore.GetResetPasswordCodeEmail(r.Context(), request.Token)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				"-",
				"-",
			})
			return
		}

		exsistingUser, err := userStore.GetUserByEmail(email)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				"-",
				"-",
			})
			return
		}

		if exsistingUser == nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				"-",
				"-",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		err = userStore.UpdateUserPassword(exsistingUser.Id, request.Password)
		if err != nil {
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				"-",
				"-",
			})
			return
		}

		_ = json.NewEncoder(w).Encode(Response{
			"success": true,
		})
	}
}


func VerifyTfa(userStore postgres.UserStoreInterface, authStore postgres.AuthStoreInterface) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

	}
}

func BypassTfa(userStore postgres.UserStoreInterface, authStore postgres.AuthStoreInterface) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

	}
}

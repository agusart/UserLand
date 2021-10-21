package auth

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"userland/api"
	"userland/api/middleware"
	"userland/store/postgres"
)

func Login(
	userStore postgres.UserStoreInterface,
	jwt middleware.JwtHandlerInterface,
	sessionStore postgres.SessionStoreInterface,
	authStore postgres.AuthStoreInterface,
	) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		loginRequest := LoginRequest{}
		_ = json.NewDecoder(r.Body).Decode(&loginRequest)

		if errMsg := loginRequest.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(Response{
				"fields": errMsg,
			})
			return
		}

		loginUser, err := userStore.GetUserByEmail(loginRequest.Email)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		if !CheckPasswordHash(loginRequest.Password, loginUser.Password){
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Code: api.ErrWringPasswordCode,
				Message: "incorrect password",
			})
			return
		}

		session, err := sessionStore.GetSessionByUserId(loginUser.Id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		if session == nil {
			session, err = sessionStore.CreateNewSession(loginUser.Id)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
				return
			}
		}


		var accessToken *middleware.JWTToken

		accessToken, err = jwt.GenerateAccessToken(loginUser.Id, session.Id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		if !loginUser.TfaEnabled {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(LoginResponse{
				RequireTfa: loginUser.TfaEnabled,
				Token: *accessToken,
			})
		}

		tfaToken, err := authStore.SendTfaVerificationCode(r.Context(), loginUser.Email, api.TfaExpiredTime)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		success := api.SendTfaEmail(loginUser.Email, tfaToken)
		if !success {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Code: "ERR_TFA",
				Message: "cant send tfa code",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(LoginResponse{
			RequireTfa: loginUser.TfaEnabled,
			Token: *accessToken,
		})
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
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				api.ErrBadRequestErrorCode,
				"please provide token",
			})
			return
		}

		email, err := authStore.GetRegistrationCodeEmail(r.Context(), token)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(err)
			return
		}

		err = userStore.VerifyUser(email)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(err)
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
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		existingUser, err := userStore.GetUserByEmail(email)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		if existingUser == nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		err = userStore.UpdateUserPassword(existingUser.Id, request.Password)
		if err != nil {
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		_ = json.NewEncoder(w).Encode(Response{
			"success": true,
		})
	}
}


func VerifyTfa(
	jwt middleware.JwtHandlerInterface,
	sessionStore postgres.SessionStoreInterface,
	) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		verifyTfaRequest := VerifyTfaRequest{}
		_ = json.NewEncoder(w).Encode(&verifyTfaRequest)

		if errMsg := verifyTfaRequest.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(Response{
				"fields": errMsg,
			})
			return
		}

		userId := r.Context().Value(api.ContextUserIdKey).(uint)
		userSession, err := sessionStore.GetSessionByUserId(userId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		tokenWithTfa, err  :=jwt.GenerateAccessTokenTfa(userId, userSession.Id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}


		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Response{
			"access_token" : tokenWithTfa,
		})
	}
}

func BypassTfa(
	jwt middleware.JwtHandlerInterface,
	sessionStore postgres.SessionStoreInterface,
	tfaStore postgres.TfaStoreInterface,
	) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bypassTfaRequest := VerifyTfaRequest{}
		_ = json.NewEncoder(w).Encode(&bypassTfaRequest)

		if errMsg := bypassTfaRequest.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(Response{
				"fields": errMsg,
			})
			return
		}

		userId := r.Context().Value(api.ContextUserIdKey).(uint)
		userSession, err := sessionStore.GetSessionByUserId(userId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		success, err := tfaStore.CheckTfaBackupCode(userId, bypassTfaRequest.Code)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		if !success {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Code : "ER-xx",
				Message: "Code is incorrect",
			})
			return
		}

		tokenWithTfa, err  :=jwt.GenerateAccessTokenTfa(userId, userSession.Id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Response{
			"access_token" : tokenWithTfa,
		})
	}
}

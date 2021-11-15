package auth

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"net/http"
	"userland/api"
	"userland/api/middleware"
	"userland/store/broker"
	"userland/store/postgres"
	"userland/store/redis"
)

func Login(
	userStore postgres.UserStoreInterface,
	jwt middleware.JwtHandlerInterface,
	sessionStore postgres.SessionStoreInterface,
	authStore redis.AuthStoreInterface,
	msgBroker broker.MessageBrokerInterface,
	) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		loginRequest := LoginRequest{}
		_ = json.NewDecoder(r.Body).Decode(&loginRequest)

		if errMsg := loginRequest.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(api.Response{
				"fields": errMsg,
			})
			return
		}

		loginUser, err := userStore.GetUserByEmail(loginRequest.Email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		if !CheckPasswordHash(loginRequest.Password, loginUser.Password){
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.ErrorResponse{
				Code: api.ErrWrongPasswordCode,
				Message: "incorrect password",
			})
			return
		}

		if !loginUser.Verified {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.ErrorResponse{
				Code: api.ErrUnverifiedCode,
				Message: "Unverified User",
			})
			return
		}

		clientName := r.Header.Get(api.ContextApiClientId)
		if clientName == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		session := postgres.Session{
			UserId: loginUser.Id,
			IP: r.RemoteAddr,
		}

		client, err := sessionStore.CreateClient(clientName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		session.Client = *client
		createdSeason, err := sessionStore.CreateNewSession(session)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		if createdSeason == nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		claim := middleware.JWTClaims{
			UserId: loginUser.Id,
			UserHasTfa: loginUser.TfaEnabled,
			SessionId: createdSeason.Id,
			TfaVerified: false,
		}

		claim.Id = uuid.NewString()
		var accessToken *middleware.JWTToken
		accessToken, err = jwt.GenerateAccessToken(claim)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		createdSeason.JwtId = accessToken.Id
		err = sessionStore.UpdateSession(*createdSeason)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		sessionCache := redis.SessionCache{
			Id : createdSeason.Id,
			UserId: createdSeason.UserId,
			JwtId: accessToken.Id,
		}

		err = authStore.InsertSessionCache(r.Context(), sessionCache)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		if !loginUser.TfaEnabled {
			err = SendLog(msgBroker, *createdSeason)
			if err != nil {
				log.Err(err)
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(LoginResponse{
				RequireTfa: loginUser.TfaEnabled,
				Token: *accessToken,
			})

			return
		}

		tfaToken, err := authStore.CreateTfaVerificationCode(r.Context(), loginUser.Email, api.TfaExpiredTime)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		success := api.SendTfaEmail(loginUser.Email, tfaToken)
		if !success {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.ErrorResponse{
				Code: "ERR-TFA",
				Message: "cant send tfa code",
			})
			return
		}

		err = SendLog(msgBroker, *createdSeason)
		if err != nil {
			log.Err(err)
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(LoginResponse{
			RequireTfa: loginUser.TfaEnabled,
			Token: *accessToken,
		})
	}
}

func Register(userStore postgres.UserStoreInterface, authStore redis.AuthStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registerRequest := RegisterRequest{}
		_ = json.NewDecoder(r.Body).Decode(&registerRequest)

		if errMsg := registerRequest.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(api.Response{
				"fields": errMsg,
			})
			return
		}

		user := NewUserFromRegisterRequest(registerRequest)
		err := userStore.RegisterUser(user)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			errResponse := api.GenerateErrorResponse(err)
			_ = json.NewEncoder(w).Encode(errResponse)
			return
		}

		token, err := authStore.CreateRegistrationVerificationCode(r.Context(), user.Email, api.VerificationExpiredTime)
		if err != nil {
			log.Error().Stack().Err(err).Msg("")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		//go func() {
		//	api.SendEmail(verificationRequest.Recipient, token)
		//}()

		log.Printf("token : %v", token)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"success": true,
		})
	}
}

func RequestVerification(userStore postgres.UserStoreInterface, authStore redis.AuthStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		verificationRequest := RegisterVerificationRequest{}
		_ = json.NewDecoder(r.Body).Decode(&verificationRequest)

		if errMsg := verificationRequest.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(api.Response{
				"fields": errMsg,
			})
			return
		}

		isVerified, err := userStore.IsUserVerified(verificationRequest.Recipient)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		if isVerified {
			_ = json.NewEncoder(w).Encode(
				api.ErrorResponse{
					Code: postgres.ErrUserAlreadyVerifiedCode,
					Message: "User Already Verified",
				},
			)
			return
		}
		//
		token, err := authStore.CreateRegistrationVerificationCode(r.Context(), verificationRequest.Recipient, api.VerificationExpiredTime)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		//go func() {
		//	api.SendEmail(verificationRequest.Recipient, token)
		//}()

		log.Printf("token : %v", token)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"success": true,
		})
	}
}

func VerifyRegister(userStore postgres.UserStoreInterface, authStore redis.AuthStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "token")
		if token == "" {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(api.ErrorResponse{
				api.ErrBadRequestErrorCode,
				"please provide token",
			})
			return
		}

		email, err := authStore.GetRegistrationCodeEmail(r.Context(), token)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}
		err = userStore.VerifyUser(email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"success": true,
		})
	}
}

func ForgetPassword(userStore postgres.UserStoreInterface, authStore redis.AuthStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		forgetPasswordRequest := ForgotPasswordRequest{}
		_ = json.NewDecoder(r.Body).Decode(&forgetPasswordRequest)

		if errMsg := forgetPasswordRequest.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(api.Response{
				"fields": errMsg,
			})
			return
		}

		existingUser, err := userStore.GetUserByEmail(forgetPasswordRequest.Email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		if existingUser == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if existingUser.DeletedAt != nil && !existingUser.Verified {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.ErrorResponse{
				api.ErrBadRequestErrorCode,
				"cant reset password",
			})
			return
		}

		token, err := authStore.CreateForgotPasswordVerificationCode(r.Context(), existingUser.Email, api.ForgotPasswordExpiredTime)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		//go func() {
		//	api.SendEmail(forgetPasswordRequest.Email, token)
		//}()

		log.Printf("token : %v", token)

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"success": true,
		})

	}
}

func ResetPassword(userStore postgres.UserStoreInterface, authStore redis.AuthStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := ResetPasswordRequest{}
		_ = json.NewDecoder(r.Body).Decode(&request)

		if errMsg := request.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(api.Response{
				"fields": errMsg,
			})
			return
		}

		email, err := authStore.GetResetPasswordCodeEmail(r.Context(), request.Token)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		existingUser, err := userStore.GetUserByEmail(email)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		if existingUser == nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}


		newPassword, err :=  HashPassword(request.Password)
		err = userStore.UpdateUserPassword(existingUser.Id, newPassword)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"success": true,
		})
	}
}


func VerifyTfa(
	jwt middleware.JwtHandlerInterface,
	userStore postgres.UserStoreInterface,
	tfaStore postgres.TfaStoreInterface,
	) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		verifyTfaRequest := VerifyTfaRequest{}
		_ = json.NewDecoder(r.Body).Decode(&verifyTfaRequest)

		if errMsg := verifyTfaRequest.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(api.Response{
				"fields": errMsg,
			})
			return
		}

		claim :=  r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)
		existingUser, err := userStore.GetUserById(claim.UserId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		if existingUser == nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		tfaDetail, err  := tfaStore.GetUserTfaDetail(claim.UserId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))

			return
		}

		success, err := api.VerifyTfaCode(tfaDetail.TfaSecret, verifyTfaRequest.Code)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))

			return
		}

		if !success {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		claim.TfaVerified = true
		tokenWithTfa, err  := jwt.GenerateAccessToken(claim)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}


		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"access_token" : tokenWithTfa,
		})
	}
}

func BypassTfa(
	jwt middleware.JwtHandlerInterface,
	tfaStore postgres.TfaStoreInterface,
	) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bypassTfaRequest := VerifyTfaRequest{}
		_ = json.NewDecoder(r.Body).Decode(&bypassTfaRequest)

		if errMsg := bypassTfaRequest.Validate(); len(errMsg) != 0 {
			w.WriteHeader(http.StatusUnprocessableEntity)
			_ = json.NewEncoder(w).Encode(api.Response{
				"fields": errMsg,
			})
			return
		}

		claim :=  r.Context().Value(api.ContextClaimsJwt).(middleware.JWTClaims)

		success, err := tfaStore.CheckTfaBackupCode(claim.UserId, bypassTfaRequest.Code)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		err = tfaStore.DeleteTfaCode(claim.UserId, bypassTfaRequest.Code)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		if !success {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.ErrorResponse{
				Code : "ER-xx",
				Message: "Code is incorrect",
			})
			return
		}

		claim.TfaVerified = true
		tokenWithTfa, err  :=jwt.GenerateAccessToken(claim)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(api.GenerateErrorResponse(err))
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(api.Response{
			"access_token" : tokenWithTfa,
		})
	}
}

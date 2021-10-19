package auth

import (
	"encoding/json"
	"net/http"
	"userland/api"
	"userland/store/postgres"
)

func Register(userStore postgres.UserStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registerRequest := RegisterRequest{}
		_ = json.NewDecoder(r.Body).Decode(&registerRequest)

		if errMsg := registerRequest.Validate(); errMsg != nil {
			_ = json.NewEncoder(w).Encode(Response{
				"fields": errMsg,
			})
			return
		}

		user := NewUserFromRegisterRequest(registerRequest)
		err := userStore.RegisterUser(user)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			errResponse := GenerateErrorResponse(err)
			_ = json.NewEncoder(w).Encode(errResponse)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Response{
			"success": true,
		})
	}
}

func Verification(userStore postgres.UserStoreInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		verificationRequest := VerificationRequest{}
		_ = json.NewDecoder(r.Body).Decode(&verificationRequest)

		if errMsg := verificationRequest.Validate(); errMsg != nil {
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
					Code: api.ErrUserAlreadyVerified,
					Message: "User Already Verified",
				},
			)
			return
		}

		token, err := userStore.SendVerification(r.Context(), verificationRequest.Recipient, api.VerificationExpiredtime)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(GenerateErrorResponse(err))
			return
		}

		go func() {
			api.SendEmail(verificationRequest.Recipient, token)
		}()

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Response{
			"success": true,
		})
	}
}
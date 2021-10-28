package api

import "time"

const BaseUrl = "192.168.99.104:8080"
const HashPasswordCost = 14

const ErrInternalServerErrorCode = "ER-500"
const ErrWrongPasswordCode = "ER-21"
const ErrUnverifiedCode = "ER-27"
const ErrBadRequestErrorCode = "ER-400"
const ErrBadRequestWrongVerifyToken = "ER-41"

const ActionVerifyEmail = "email.verify"

const VerificationExpiredTime = 15 * time.Minute
const TfaExpiredTime = 5 * time.Minute
const ForgotPasswordExpiredTime = 3 * time.Hour

const ContextClaimsJwt = "context.jwt"
const ContextApiClientId = "X-API-ClientID"

const RefreshTokenExpTime = 30 * 24 * time.Hour
const AccessTokenExpTime = 50 * time.Minute


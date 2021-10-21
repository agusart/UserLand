package api

import "time"

const BaseUrl = "192.168.99.104:8080"
const HashPasswordCost = 14

const ErrInternalServerErrorCode = "ER-1"

const ErrWringPasswordCode = "ER-21"

const ErrBadRequestErrorCode = "ER-4"
const ErrBadRequestWrongVerifyToken = "ER-41"



const ActionVerifyEmail = "email.verify"
const VerificationExpiredTime = 15 * time.Minute
const TfaExpiredTime = 5 * time.Minute


const ForgotPasswordExpiredTime = 3 * time.Hour


const ContextUserIdKey = "context.userId"
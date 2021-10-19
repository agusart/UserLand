package api

import "time"

const BaseUrl = "192.168.99.104:8080"

const HashPasswordCost = 14
const ErrInternalServerError = "ER-1"
const ErrUserAlreadyRegistered = "ER-21"
const ErrUserAlreadyVerified = "ER-21"



const ActionVerifyEmail = "verify.email"
const VerificationExpiredtime = 15 * time.Minute
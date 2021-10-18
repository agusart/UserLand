package api

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

func SendEmail(to []string, msg []byte) bool {
	auth := smtp.PlainAuth("", os.Getenv("SMTP_LOGIN"), os.Getenv("SMTP_PASSWORD"), os.Getenv("SMTP_HOST"))
	smtpAddress := fmt.Sprintf("%s:%v", os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT"))

	err := smtp.SendMail(smtpAddress, auth, os.Getenv("SMTP_LOGIN"), to, msg)

	if err != nil {
		log.Fatal(err)
		return false
	}

	return true
}


package api

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

func SendEmail(to string, token string) bool {
	auth := smtp.PlainAuth("", os.Getenv("SMTP_LOGIN"), os.Getenv("SMTP_PASSWORD"), os.Getenv("SMTP_HOST"))
	smtpAddress := fmt.Sprintf("%s:%v", os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT"))
	msg := []byte("To: "+ to + "\r\n" +
		"From: hello@userland.dev\r\n" +
		"Subject: Hello!\r\n" +
		"\r\n" +
		"This is your confirmation link " + BaseUrl + "/verify/" + token )

	err := smtp.SendMail(smtpAddress, auth, os.Getenv("SMTP_LOGIN"), []string{to}, msg)

	if err != nil {
		log.Fatal(err)
		return false
	}

	return true
}


package api

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

func SendRegistrationEmail(to, token string) bool{
	msg := []byte("To: "+ to + "\r\n" +
		"From: hello@userland.dev\r\n" +
		"Subject: Hello!\r\n" +
		"\r\n" +
		"This is your registration confirmation link " + BaseUrl + "/verify/" + token )

	return sendEmail(msg, to)
}

func SendTfaEmail(to, token string) bool{
	msg := []byte("To: "+ to + "\r\n" +
		"From: hello@userland.dev\r\n" +
		"Subject: Hello!\r\n" +
		"\r\n" +
		"This is your Tfa code " +  token )

	return sendEmail(msg, to)
}


func SendForgotPasswordEmail(to, token string) bool{
	msg := []byte("To: "+ to + "\r\n" +
		"From: hello@userland.dev\r\n" +
		"Subject: Hello!\r\n" +
		"\r\n" +
		"This is your reset password confirmation link " + BaseUrl + "/verify/" + token )

	return sendEmail(msg, to)
}


func sendEmail(msg []byte, to string) bool {
	auth := smtp.PlainAuth("", os.Getenv("SMTP_LOGIN"), os.Getenv("SMTP_PASSWORD"), os.Getenv("SMTP_HOST"))
	smtpAddress := fmt.Sprintf("%s:%v", os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT"))

	err := smtp.SendMail(smtpAddress, auth, os.Getenv("SMTP_LOGIN"), []string{to}, msg)

	if err != nil {
		log.Fatal(err)
		return false
	}

	return true
}


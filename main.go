package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"userland/api"
)


func main(){
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
	to := []string{"agus.kbk@gmail.com"}

	// Create a message and convert it into bytes
	msg := []byte("This is the email is sent using golang and sendinblue.\r\n")

	status := api.SendEmail(to, msg)

	if status {
		fmt.Printf("Email sent successfully.")
	} else {
		fmt.Printf("Email sent failed.")
	}
}
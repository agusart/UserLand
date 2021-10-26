package me

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"github.com/dgryski/dgoogauth"
	"github.com/skip2/go-qrcode"
	"image"
	"image/png"
	"strings"
)

func randStr(strSize int, randType string) string {

	var dictionary string

	if randType == "alphanum" {
		dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	if randType == "alpha" {
		dictionary = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	}

	if randType == "number" {
		dictionary = "0123456789"
	}

	var bytesData = make([]byte, strSize)
	rand.Read(bytesData)
	for k, v := range bytesData {
		bytesData[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytesData)
}

func generateTfaSecret() string {
	randomStr := randStr(6, "alphanum")
	secret := base32.StdEncoding.EncodeToString([]byte(randomStr))

	return secret
}

func generateAuthLink(secret, issuer string) string {
	authLink := "otpauth://totp/Userland?secret=" + secret + "&issuer=" + issuer
	return authLink
}

func getBase64String(m image.Image) string {
	var buf bytes.Buffer
	err := png.Encode(&buf, m)
	if err != nil {
		panic(err)
	}
	enc := base64.StdEncoding.EncodeToString(buf.Bytes())

	return "data:image/png;base64," + enc
}

func GenerateQRString(authLink string) (string, error) {
	code, err := qrcode.New(authLink, qrcode.Low)
	if err != nil {
		return "", err
	}

	return getBase64String(code.Image(256)), nil
}

func VerifyTfaCode(request ActivateTfaRequest) (bool, error) {
	otpConfig := &dgoogauth.OTPConfig{
		Secret:      strings.TrimSpace(request.Secret),
		WindowSize:  3,
		HotpCounter: 0,
	}

	trimmedToken := strings.TrimSpace(request.Code)
	ok, err := otpConfig.Authenticate(trimmedToken)

	return ok, err
}

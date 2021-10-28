package me

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"encoding/base64"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/skip2/go-qrcode"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"userland/store/postgres"
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


type FileHelperInterface interface {
	ReadFile(filePath string)  ([]byte, error)
	Create(name string) (*os.File, error)
	Copy(dst io.Writer, src io.Reader) error
	IsAllowedContentType(file multipart.File) (bool, error)
}

type FileHelper struct {}

func (f FileHelper) IsAllowedContentType(file multipart.File) (bool, error) {
	fileHeader := make([]byte, 512)
	_, _ = file.Read(fileHeader)

	filTipe := http.DetectContentType(fileHeader)

	if !IsValidImage(filTipe) {
		return false, errors.New("ivalid img type")
	}

	return  IsValidImageDimension(file)
}

func (f FileHelper) ReadFile(filePath string) ([]byte, error) {
	return ioutil.ReadFile(filePath)
}

func (f FileHelper) Create(name string) (*os.File, error) {
	suffix := uuid.NewString()
	photoName := suffix + name

	return os.Create(postgres.PhotoPath + "/" + photoName)
}

func (f FileHelper) Copy(dst io.Writer, src io.Reader) error {
	_, err :=  io.Copy(dst, src)
	return  err
}


func IsValidImage(category string) bool {
	switch category {
	case
		"image/jpeg":
		return true
	}

	return false
}


func IsValidImageDimension(f multipart.File) (bool, error){
	defer f.Close()
	im, _, err := image.DecodeConfig(f)
	if err != nil {
		return false, err
	}

	return im.Width == im.Height &&
		(im.Width > 200 && im.Width < 500) &&
		(im.Height > 200 && im.Height < 500), nil
}
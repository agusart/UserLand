package me

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGenerateQRString(t *testing.T) {
	secret := generateTfaSecret()
	fmt.Println(secret)
	link := generateAuthLink(secret, "user")
	res, err := GenerateQRString(link)
	assert.NoError(t, err)
	fmt.Println(res)
}

//func TestVerifyQR(t *testing.T) {
//	ok, err := VerifyTfaCode(ActivateTfaRequest{
//		Secret: "MY3WI6LHNI======",
//		Code: "512287",
//	})
//
//	assert.NoError(t, err)
//	assert.True(t, ok)
//}

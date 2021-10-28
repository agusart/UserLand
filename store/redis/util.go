package redis

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
)

func StructToArgs(s interface{}) []interface{} {
	m := getMapFromStruct(s)
	var all []interface{}

	for k, v := range m {
		if v == "" {
			continue
		}
		all = append(all, k)
		all = append(all, v)
	}

	return all
}

func getMapFromStruct(s interface{}) map[string]interface{} {
	var inInterface map[string]interface{}
	res, err := json.Marshal(s)
	if err != nil {
		return nil
	}

	err = json.Unmarshal(res, &inInterface)
	if err != nil {
		return nil
	}

	return inInterface
}

func GenerateUserSessionKey(userId, sessionId uint) string{
	return fmt.Sprintf("%d:session:%d", userId, sessionId)
}

func tokenGenerator() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}




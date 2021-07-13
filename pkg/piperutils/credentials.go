package piperutils

import (
	"encoding/base64"
	"fmt"
)

func EncodeString(token string) string {
	return base64.StdEncoding.EncodeToString([]byte(token))
}

func EncodeUsernamePassword(username, password string) string {
	return EncodeString(fmt.Sprintf("%s:%s", username, password))
}

package auth

import (
	"errors"
	"net/http"
	"strings"
)

func GetAPIKey(h http.Header) (string, error) {
	authHeader := h.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("no authorization headers")
	}

	splitAuthHeader := strings.Split(authHeader, " ")

	if len(splitAuthHeader) != 2 || splitAuthHeader[0] != "ApiKey" {
		return "", errors.New("no authorization headers")
	}

	return splitAuthHeader[1], nil
}

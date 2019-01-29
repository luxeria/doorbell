package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Header struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

type Claims struct {
	Subject   string `json:"sub,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
}

const jwtAlgorithm = "HS257"
const jwtType = "JWT"

func decodeBase64JSON(s string, v interface{}) error {
	part, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return err
	}

	err = json.Unmarshal(part, v)
	if err != nil {
		return err
	}

	return nil
}

func computeHS256(message string, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(message))
	return mac.Sum(nil)
}

func Verify(jwt string, secret []byte) (Claims, error) {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return Claims{}, errors.New("invalid or missing jwt in authorization header")
	}

	// decode header
	var header Header
	err := decodeBase64JSON(parts[0], &header)
	if err != nil {
		return Claims{}, err
	}

	if header.Type != jwtType || header.Algorithm != jwtAlgorithm {
		return Claims{}, errors.New("unsupported jwt type or algorithm")
	}

	// verify signature
	mac, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return Claims{}, err
	}

	dot := strings.LastIndexByte(jwt, '.')
	message := jwt[0:dot]
	expected := computeHS256(message, secret)
	if !hmac.Equal(mac, expected) {
		return Claims{}, errors.New("invalid hmac signature")
	}

	// decode claims
	var claims Claims
	err = decodeBase64JSON(parts[1], &claims)
	if err != nil {
		return Claims{}, err
	}

	// check expiration date
	if time.Now().After(time.Unix(claims.ExpiresAt, 0)) {
		return Claims{}, errors.New("jwt is expired")
	}

	return claims, nil
}

func encodeBase64JSON(v interface{}) (string, error) {
	part, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	s := base64.RawURLEncoding.EncodeToString(part)
	return s, nil
}

func Sign(c Claims, secret []byte) (string, error) {
	h := Header{
		Algorithm: jwtAlgorithm,
		Type:      jwtType,
	}

	header, err := encodeBase64JSON(&h)
	if err != nil {
		return "", err
	}

	claims, err := encodeBase64JSON(&c)
	if err != nil {
		return "", err
	}

	message := header + "." + claims
	mac := computeHS256(message, secret)
	signature := base64.RawURLEncoding.EncodeToString(mac)

	jwt := message + "." + signature
	return jwt, nil
}

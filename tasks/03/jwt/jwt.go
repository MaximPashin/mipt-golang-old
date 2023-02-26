package jwt

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"hash"
	"time"
)

type SignMethod string

const (
	HS256 SignMethod = "HS256"
	HS512 SignMethod = "HS512"
)

var (
	ErrInvalidSignMethod      = errors.New("invalid sign method")
	ErrSignatureInvalid       = errors.New("signature invalid")
	ErrTokenExpired           = errors.New("token expired")
	ErrSignMethodMismatched   = errors.New("sign method mismatched")
	ErrConfigurationMalformed = errors.New("configuration malformed")
	ErrInvalidToken           = errors.New("invalid token")
)

func Encode(data interface{}, opts ...Option) ([]byte, error) {
	conf := config{}
	for _, opt := range opts {
		opt(&conf)
	}
	expires, err := getExpires(&conf)
	if err != nil {
		return nil, err
	}

	key := conf.Key
	signMethod := conf.SignMethod
	var tokenHashFunc func() hash.Hash
	switch signMethod {
	case HS256:
		tokenHashFunc = sha256.New
	case HS512:
		tokenHashFunc = sha512.New
	default:
		return nil, ErrInvalidSignMethod
	}

	header := make(map[string]interface{})
	header["alg"] = signMethod
	header["typ"] = "JWT"
	payload := make(map[string]interface{})
	payload["d"] = data
	if expires != nil {
		payload["exp"] = expires.Unix()
	}

	encoder := base64.RawURLEncoding
	var token bytes.Buffer

	marshaledHeader, _ := json.Marshal(header)
	marshaledPayload, _ := json.Marshal(payload)

	token.WriteString(encoder.EncodeToString(marshaledHeader))
	token.WriteString(".")
	token.WriteString(encoder.EncodeToString(marshaledPayload))

	tokenHash := hmac.New(tokenHashFunc, key)
	tokenHash.Write(token.Bytes())
	vrfSign := tokenHash.Sum(nil)
	token.WriteString(".")
	token.WriteString(encoder.EncodeToString(vrfSign))
	return token.Bytes(), nil
}

func Decode(token []byte, data interface{}, opts ...Option) error {
	// TODO: implement me
	return nil
}

func getExpires(conf *config) (*time.Time, error) {
	expires := conf.Expires
	ttl := conf.TTL
	if expires != nil && ttl != nil {
		return nil, ErrConfigurationMalformed
	}
	if ttl != nil {
		temp := timeFunc().Add(*ttl)
		expires = &temp
	}
	if expires == nil {
		return nil, nil
	}
	if timeFunc().After(*expires) {
		return nil, ErrConfigurationMalformed
	}
	return expires, nil
}

// To mock time in tests
var timeFunc = time.Now

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
	conf := getConfig(opts...)

	header, err := getHeader(conf)
	if err != nil {
		return nil, err
	}

	payload, err := getPayload(&data, conf)
	if err != nil {
		return nil, err
	}

	key := conf.Key

	token := buildToken(header, payload, conf.SignMethod, key)
	return token, nil
}

func Decode(token []byte, data interface{}, opts ...Option) error {
	conf := getConfig(opts...)

	tokenParts := bytes.Split(token, []byte("."))
	if len(tokenParts) != 3 {
		return ErrInvalidToken
	}
	tokenHeader := tokenParts[0]
	tokenPayload := tokenParts[1]

	header, err := decodeHeader(tokenHeader)
	if err != nil {
		return err
	}

	payload, err := decodePayload(tokenPayload)
	if err != nil {
		return err
	}

	key := conf.Key

	signMethod, err := getSignMethod(header)

	if signMethod != conf.SignMethod {
		return ErrSignMethodMismatched
	}

	buildedToken := buildToken(header, payload, signMethod, key)

	if bytes.Compare(buildedToken, token) != 0 {
		return ErrSignatureInvalid
	}

	exp_data, ok := (*payload)["exp"]
	if ok {
		expires := time.Unix(0, int64(exp_data.(float64)))
		if timeFunc().After(expires) {
			return ErrTokenExpired
		}
	}

	switch v := data.(type) {
	case *map[string]interface{}:
		*v = (*payload)["d"].(map[string]interface{})
	case *interface{}:
		*v = (*payload)["d"].(interface{})
	}

	return nil
}

func getConfig(opts ...Option) *config {
	conf := config{}
	for _, opt := range opts {
		opt(&conf)
	}
	return &conf
}

func getHeader(conf *config) (*map[string]interface{}, error) {
	signMethod := conf.SignMethod
	if signMethod != HS256 && signMethod != HS512 {
		return nil, ErrInvalidSignMethod
	}
	header := map[string]interface{}{
		"alg": signMethod,
		"typ": "JWT",
	}
	return &header, nil
}

func getPayload(data *interface{}, conf *config) (*map[string]interface{}, error) {
	expires, err := getExpires(conf)
	if err != nil {
		return nil, err
	}

	payload := map[string]interface{}{
		"d": *data,
	}
	if expires != nil {
		payload["exp"] = expires.Unix()
	}
	return &payload, nil
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
		return expires, ErrConfigurationMalformed
	}
	return expires, nil
}

func buildToken(header *map[string]interface{}, payload *map[string]interface{}, signMethod SignMethod, key []byte) []byte {
	encoder := base64.RawURLEncoding
	var token bytes.Buffer

	marshaledHeader, _ := json.Marshal(header)
	marshaledPayload, _ := json.Marshal(payload)

	token.WriteString(encoder.EncodeToString(marshaledHeader))
	token.WriteString(".")
	token.WriteString(encoder.EncodeToString(marshaledPayload))

	tokenHashFunc := getHashFunc(signMethod)

	tokenHash := hmac.New(tokenHashFunc, key)
	tokenHash.Write(token.Bytes())
	vrfSign := tokenHash.Sum(nil)
	token.WriteString(".")
	token.WriteString(encoder.EncodeToString(vrfSign))
	return token.Bytes()
}

func getHashFunc(signMethod SignMethod) func() hash.Hash {
	var tokenHashFunc func() hash.Hash
	switch signMethod {
	case HS256:
		tokenHashFunc = sha256.New
	case HS512:
		tokenHashFunc = sha512.New
	}
	return tokenHashFunc
}

func decodeHeader(tokenHeader []byte) (*map[string]interface{}, error) {
	encoder := base64.RawURLEncoding

	decodedHeader := make([]byte, encoder.DecodedLen(len(tokenHeader)))
	_, err := encoder.Decode(decodedHeader, tokenHeader)
	if err != nil {
		return nil, ErrInvalidToken
	}
	header := make(map[string]interface{})
	err = json.Unmarshal(decodedHeader, &header)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if header["typ"] != "JWT" || len(header) != 2 {
		return nil, ErrInvalidToken
	}

	return &header, nil
}

func decodePayload(tokenPayload []byte) (*map[string]interface{}, error) {
	encoder := base64.RawURLEncoding

	decodedPayload := make([]byte, encoder.DecodedLen(len(tokenPayload)))
	_, err := encoder.Decode(decodedPayload, tokenPayload)
	if err != nil {
		return nil, ErrInvalidToken
	}

	payload := make(map[string]interface{})
	err = json.Unmarshal(decodedPayload, &payload)
	if err != nil {
		return nil, ErrInvalidToken
	}
	return &payload, nil
}

func getSignMethod(header *map[string]interface{}) (SignMethod, error) {
	var signMethod SignMethod
	switch v := (*header)["alg"].(type) {
	case string:
		signMethod = SignMethod(v)
	default:
		return "", ErrInvalidToken
	}

	if signMethod != HS256 && signMethod != HS512 {
		return "", ErrInvalidSignMethod
	}

	return signMethod, nil
}

// To mock time in tests
var timeFunc = time.Now

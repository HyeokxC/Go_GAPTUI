package exchanges

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hyeokx/Go_GAPTUI/internal/config"
)

func currentMillis() string {
	return fmt.Sprintf("%d", time.Now().UnixMilli())
}

func BinanceSignedQuery(extraParams string, keys *config.ApiKeyPair) (string, error) {
	if err := validateKeys(keys); err != nil {
		return "", err
	}

	timestamp := currentMillis()
	query := "timestamp=" + timestamp
	if extraParams != "" {
		query = extraParams + "&" + query
	}

	signature := hmacSHA256Hex(keys.SecretKey, query)
	return query + "&signature=" + signature, nil
}

func UpbitJWTToken(keys *config.ApiKeyPair) (string, error) {
	if err := validateKeys(keys); err != nil {
		return "", err
	}

	header := base64RawURLEncode(`{"alg":"HS256","typ":"JWT"}`)
	claims := base64RawURLEncode(fmt.Sprintf(`{"access_key":"%s","nonce":"%s"}`, keys.ApiKey, uuid.NewString()))
	signingInput := header + "." + claims
	signature := hmacSHA256Base64RawURL(keys.SecretKey, signingInput)

	return signingInput + "." + signature, nil
}

func UpbitJWTWithQuery(keys *config.ApiKeyPair, queryString string) (string, error) {
	if err := validateKeys(keys); err != nil {
		return "", err
	}

	queryHash := sha512Hex(queryString)
	header := base64RawURLEncode(`{"alg":"HS256","typ":"JWT"}`)
	claims := base64RawURLEncode(fmt.Sprintf(`{"access_key":"%s","nonce":"%s","query_hash":"%s","query_hash_alg":"SHA512"}`,
		keys.ApiKey,
		uuid.NewString(),
		queryHash,
	))
	signingInput := header + "." + claims
	signature := hmacSHA256Base64RawURL(keys.SecretKey, signingInput)

	return signingInput + "." + signature, nil
}

func BithumbJWTToken(keys *config.ApiKeyPair) (string, error) {
	if err := validateKeys(keys); err != nil {
		return "", err
	}

	timestamp := time.Now().UnixMilli()
	header := base64RawURLEncode(`{"alg":"HS256","typ":"JWT"}`)
	claims := base64RawURLEncode(fmt.Sprintf(`{"access_key":"%s","nonce":"%s","timestamp":%d}`,
		keys.ApiKey,
		uuid.NewString(),
		timestamp,
	))
	signingInput := header + "." + claims
	signature := hmacSHA256Base64RawURL(keys.SecretKey, signingInput)

	return signingInput + "." + signature, nil
}

func BithumbJWTWithQuery(keys *config.ApiKeyPair, queryString string) (string, error) {
	if err := validateKeys(keys); err != nil {
		return "", err
	}

	timestamp := time.Now().UnixMilli()
	queryHash := sha512Hex(queryString)
	header := base64RawURLEncode(`{"alg":"HS256","typ":"JWT"}`)
	claims := base64RawURLEncode(fmt.Sprintf(`{"access_key":"%s","nonce":"%s","timestamp":%d,"query_hash":"%s","query_hash_alg":"SHA512"}`,
		keys.ApiKey,
		uuid.NewString(),
		timestamp,
		queryHash,
	))
	signingInput := header + "." + claims
	signature := hmacSHA256Base64RawURL(keys.SecretKey, signingInput)

	return signingInput + "." + signature, nil
}

func BybitSign(method string, queryParams string, body string, keys *config.ApiKeyPair) (signature string, timestamp string, err error) {
	if err = validateKeys(keys); err != nil {
		return "", "", err
	}

	const recvWindow = "20000"
	timestamp = currentMillis()
	payloadData := body
	if strings.EqualFold(method, "GET") {
		payloadData = queryParams
	}
	payload := timestamp + keys.ApiKey + recvWindow + payloadData

	signature = hmacSHA256Hex(keys.SecretKey, payload)
	return signature, timestamp, nil
}

func BitgetSign(method string, requestPath string, body string, keys *config.ApiKeyPair) (signature string, timestamp string, err error) {
	if err = validateKeys(keys); err != nil {
		return "", "", err
	}

	timestamp = currentMillis()
	preSign := timestamp + strings.ToUpper(method) + requestPath + body
	signature = hmacSHA256Base64Std(keys.SecretKey, preSign)

	return signature, timestamp, nil
}

func OKXSign(method string, requestPath string, body string, keys *config.ApiKeyPair) (signature string, timestamp string, err error) {
	if err = validateKeys(keys); err != nil {
		return "", "", err
	}

	timestamp = time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	preSign := timestamp + strings.ToUpper(method) + requestPath + body
	signature = hmacSHA256Base64Std(keys.SecretKey, preSign)

	return signature, timestamp, nil
}

func GateSign(method string, path string, query string, body string, keys *config.ApiKeyPair) (signature string, timestamp string, err error) {
	if err = validateKeys(keys); err != nil {
		return "", "", err
	}

	timestamp = fmt.Sprintf("%d", time.Now().Unix())
	bodyHash := sha512Hex(body)
	preSign := strings.ToUpper(method) + "\n" + path + "\n" + query + "\n" + bodyHash + "\n" + timestamp
	signature = hmacSHA512Hex(keys.SecretKey, preSign)

	return signature, timestamp, nil
}

func validateKeys(keys *config.ApiKeyPair) error {
	if keys == nil {
		return fmt.Errorf("api keys are nil")
	}
	if keys.ApiKey == "" {
		return fmt.Errorf("api key is empty")
	}
	if keys.SecretKey == "" {
		return fmt.Errorf("secret key is empty")
	}

	return nil
}

func base64RawURLEncode(value string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(value))
}

func sha512Hex(value string) string {
	sum := sha512.Sum512([]byte(value))
	return hex.EncodeToString(sum[:])
}

func hmacSHA256Hex(secret string, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func hmacSHA512Hex(secret string, payload string) string {
	mac := hmac.New(sha512.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func hmacSHA256Base64RawURL(secret string, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func hmacSHA256Base64Std(secret string, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

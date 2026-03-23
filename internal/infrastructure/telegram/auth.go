package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// TgUser represents the user JSON embedded inside initData.
type TgUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// ValidateInitData verifies the Telegram WebApp initData string and extracts the user.
func ValidateInitData(initData, botToken string) (*TgUser, error) {
	// 1. Parse the URL-encoded initData
	parsedData, err := url.ParseQuery(initData)
	if err != nil {
		return nil, errors.New("invalid initData format")
	}

	// 2. Extract the hash signature provided by Telegram
	hash := parsedData.Get("hash")
	if hash == "" {
		return nil, errors.New("missing hash in initData")
	}
	parsedData.Del("hash") // Remove hash before building the data-check-string

	// 3. Sort the remaining key-value pairs alphabetically
	var keys []string
	for k := range parsedData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 4. Build the data-check-string (k1=v1\nk2=v2...)
	var dataCheckArr []string
	for _, k := range keys {
		dataCheckArr = append(dataCheckArr, fmt.Sprintf("%s=%s", k, parsedData.Get(k)))
	}
	dataCheckString := strings.Join(dataCheckArr, "\n")

	// 5. Generate the secret key (HMAC256 of botToken with key "WebAppData")
	secretKeyHmac := hmac.New(sha256.New, []byte("WebAppData"))
	secretKeyHmac.Write([]byte(botToken))
	secretKey := secretKeyHmac.Sum(nil)

	// 6. Calculate our own signature
	signatureHmac := hmac.New(sha256.New, secretKey)
	signatureHmac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(signatureHmac.Sum(nil))

	// 7. Compare signatures
	if hash != expectedHash {
		return nil, errors.New("invalid telegram signature (nice try, hacker)")
	}

	// 8. Extract User JSON
	userStr := parsedData.Get("user")
	if userStr == "" {
		return nil, errors.New("missing user data")
	}

	var tgUser TgUser
	if err := json.Unmarshal([]byte(userStr), &tgUser); err != nil {
		return nil, fmt.Errorf("failed to parse user json: %w", err)
	}

	return &tgUser, nil
}

func ValidateInitDataWithTokens(initData string, botTokens ...string) (*TgUser, error) {
	var lastErr error

	for _, botToken := range botTokens {
		if botToken == "" {
			continue
		}

		tgUser, err := ValidateInitData(initData, botToken)
		if err == nil {
			return tgUser, nil
		}

		lastErr = err
	}

	if lastErr == nil {
		lastErr = errors.New("no bot tokens configured")
	}

	return nil, lastErr
}

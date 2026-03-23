package telegram

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type webhookResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
	Result      struct {
		URL string `json:"url"`
	} `json:"result"`
}

func EnsureWebhook(botToken string, webhookURL string) error {
	if botToken == "" {
		return fmt.Errorf("telegram bot token is not configured")
	}
	if webhookURL == "" {
		return nil
	}

	client := &http.Client{Timeout: 10 * time.Second}

	currentURL, err := getWebhookURL(client, botToken)
	if err != nil {
		return err
	}
	if currentURL == webhookURL {
		return nil
	}

	return setWebhookURL(client, botToken, webhookURL)
}

func getWebhookURL(client *http.Client, botToken string) (string, error) {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/getWebhookInfo", botToken)

	resp, err := client.Get(endpoint)
	if err != nil {
		return "", fmt.Errorf("telegram getWebhookInfo request failed: %w", err)
	}
	defer resp.Body.Close()

	var payload webhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("telegram getWebhookInfo decode failed: %w", err)
	}
	if !payload.OK {
		if payload.Description == "" {
			payload.Description = "telegram getWebhookInfo returned not ok"
		}
		return "", errors.New(payload.Description)
	}

	return payload.Result.URL, nil
}

func setWebhookURL(client *http.Client, botToken string, webhookURL string) error {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/setWebhook", botToken)

	form := url.Values{}
	form.Set("url", webhookURL)

	resp, err := client.PostForm(endpoint, form)
	if err != nil {
		return fmt.Errorf("telegram setWebhook request failed: %w", err)
	}
	defer resp.Body.Close()

	var payload webhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("telegram setWebhook decode failed: %w", err)
	}
	if !payload.OK {
		if payload.Description == "" {
			payload.Description = "telegram setWebhook returned not ok"
		}
		return errors.New(payload.Description)
	}

	return nil
}

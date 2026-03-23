package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type Notifier interface {
	SendAlert(message string)
	Start(ctx context.Context)
}

type Sender interface {
	SendMessage(chatID int64, message string) (int64, error)
	SendReplyableMessage(chatID int64, message string) (int64, error)
	SendHTMLMessage(chatID int64, message string) (int64, error)
}

type botNotifier struct {
	logger *slog.Logger
	queue  chan string // Канал (очередь) для сообщений
	token  string
	client *http.Client
}

// NewNotifier creates a new asynchronous Telegram notifier.
func NewNotifier(logger *slog.Logger, token string) Notifier {
	return &botNotifier{
		logger: logger,
		queue:  make(chan string, 100), // Буферизованный канал на 100 сообщений
		token:  token,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func NewSender(logger *slog.Logger, token string) Sender {
	return &botNotifier{
		logger: logger,
		queue:  make(chan string, 1),
		token:  token,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// SendAlert pushes a message to the queue without blocking the main execution.
func (b *botNotifier) SendAlert(message string) {
	select {
	case b.queue <- message:
		b.logger.Debug("message queued for telegram")
	default:
		// Если канал переполнен (все 100 слотов заняты), мы просто логируем ошибку,
		// но НЕ "роняем" основной поток клиента. Это часть Resilience (отказоустойчивости).
		b.logger.Warn("telegram alert queue is full, dropping message", slog.String("msg", message))
	}
}

func (b *botNotifier) SendMessage(chatID int64, message string) (int64, error) {
	return b.sendMessage(chatID, message, false, "")
}

func (b *botNotifier) SendReplyableMessage(chatID int64, message string) (int64, error) {
	return b.sendMessage(chatID, message, true, "")
}

func (b *botNotifier) SendHTMLMessage(chatID int64, message string) (int64, error) {
	return b.sendMessage(chatID, message, false, "HTML")
}

func (b *botNotifier) sendMessage(chatID int64, message string, forceReply bool, parseMode string) (int64, error) {
	if chatID == 0 {
		return 0, fmt.Errorf("telegram chat id is required")
	}
	if message == "" {
		return 0, fmt.Errorf("message is required")
	}
	if b.token == "" {
		return 0, fmt.Errorf("telegram bot token is not configured")
	}

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", b.token)
	form := url.Values{}
	form.Set("chat_id", fmt.Sprintf("%d", chatID))
	form.Set("text", message)
	if parseMode != "" {
		form.Set("parse_mode", parseMode)
	}
	if forceReply {
		form.Set("reply_markup", `{"force_reply":true,"input_field_placeholder":"Відповісти на повідомлення"}`)
	} else if parseMode == "HTML" {
		form.Set("disable_web_page_preview", "true")
	}

	resp, err := b.client.PostForm(endpoint, form)
	if err != nil {
		return 0, fmt.Errorf("telegram send request failed: %w", err)
	}
	defer resp.Body.Close()

	var payload struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
		Result      struct {
			MessageID int64 `json:"message_id"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return 0, fmt.Errorf("telegram response decode failed: %w", err)
	}
	if !payload.OK {
		if payload.Description == "" {
			payload.Description = "telegram api returned not ok"
		}
		return 0, errors.New(payload.Description)
	}

	return payload.Result.MessageID, nil
}

// Start runs a background worker (Goroutine) that listens to the queue.
func (b *botNotifier) Start(ctx context.Context) {
	b.logger.Info("starting telegram notifier worker")

	go func() {
		for {
			select {
			case <-ctx.Done():
				b.logger.Info("stopping telegram notifier worker")
				return
			case msg := <-b.queue:
				// Имитация задержки сети (обращение к API Telegram)
				time.Sleep(1 * time.Second)

				b.logger.Info("🚀 [TELEGRAM MSG SENT]: " + msg)
			}
		}
	}()
}

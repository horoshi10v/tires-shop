package telegram

import (
	"context"
	"log/slog"
	"time"
)

// Notifier defines the interface for sending alerts.
type Notifier interface {
	SendAlert(message string)
	Start(ctx context.Context) // Starts the background worker
}

type botNotifier struct {
	logger *slog.Logger
	queue  chan string // Канал (очередь) для сообщений
	// token  string    // В будущем тут будет токен бота
	// chatID string    // И ID чата директора
}

// NewNotifier creates a new asynchronous Telegram notifier.
func NewNotifier(logger *slog.Logger) Notifier {
	return &botNotifier{
		logger: logger,
		queue:  make(chan string, 100), // Буферизованный канал на 100 сообщений
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

				// В реальном проекте здесь будет:
				// http.Post("https://api.telegram.org/bot<TOKEN>/sendMessage?chat_id=<ID>&text=" + msg)

				b.logger.Info("🚀 [TELEGRAM MSG SENT]: " + msg)
			}
		}
	}()
}

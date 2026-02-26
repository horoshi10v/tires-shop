package qrcode

import (
	"fmt"

	qr "github.com/skip2/go-qrcode"
)

// Generator defines the contract for creating QR codes.
type Generator interface {
	GeneratePNG(data string, size int) ([]byte, error)
}

type qrGenerator struct{}

func NewQRGenerator() Generator {
	return &qrGenerator{}
}

// GeneratePNG creates a PNG image buffer from the given string.
func (g *qrGenerator) GeneratePNG(data string, size int) ([]byte, error) {
	// Level M (Medium) recovers up to 15% of data if the sticker is dirty/scratched
	pngBytes, err := qr.Encode(data, qr.Medium, size)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code: %w", err)
	}
	return pngBytes, nil
}

package models

// User represents a person interacting with the system via Telegram.
type User struct {
	Base
	TelegramID  int64  `gorm:"uniqueIndex;not null"` // Unique Telegram ID
	Username    string `gorm:"index"`                // @username (can be empty if user hid it)
	FirstName   string `gorm:"type:varchar(100)"`
	LastName    string `gorm:"type:varchar(100)"`
	PhoneNumber string `gorm:"type:varchar(20);index"`                 // Added for search/contact
	Role        string `gorm:"type:varchar(20);default:'BUYER';index"` // BUYER, STAFF, ADMIN
}

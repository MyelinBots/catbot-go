package user

import (
	"time"

	"github.com/MyelinBots/catbot-go/internal/services/lovemeter"
	"gorm.io/gorm"
)

type User struct {
	// Explicit fields so you control JSON keys precisely.
	ID        uint           `gorm:"primaryKey"       json:"id"`
	CreatedAt time.Time      `                        json:"created_at"`
	UpdatedAt time.Time      `                        json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"            json:"-"`

	Nickname  string              `gorm:"size:100;not null;uniqueIndex"                 json:"nickname"`
	LoveMeter lovemeter.LoveMeter `gorm:"embedded;embeddedPrefix:love_"                json:"lovemeter"`
}

// Optional: set explicit table name
func (User) TableName() string {
	return "users"
}

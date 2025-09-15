package cat_player

import (
	"time"

	"gorm.io/gorm"
)

type CatPlayer struct {
	gorm.Model `json:"-"`
	ID         string    `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name       string    `gorm:"column:name;type:text;not null" json:"name"`
	LoveMeter  int       `gorm:"column:love_meter;type:int;not null" json:"love_meter"`
	Count      int       `gorm:"column:count;type:int;not null" json:"count"`
	Network    string    `gorm:"column:network;type:text;not null" json:"network"`
	Channel    string    `gorm:"column:channel;type:text;not null" json:"channel"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// set table name
func (CatPlayer) TableName() string {
	return "cat_player"
}

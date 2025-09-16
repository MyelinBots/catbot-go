// internal/user/model.go
package user

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Nickname  string `gorm:"uniqueIndex;size:64;not null"`
	LoveScore int    `gorm:"default:0"`
}

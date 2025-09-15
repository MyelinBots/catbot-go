package user

import (
	"time"

	"gorm.io/gorm"
)

// ----- Model -----

type UserWithLove struct {
	ID        uint   `gorm:"primaryKey"`
	Nickname  string `gorm:"uniqueIndex;size:64;not null"`
	LoveScore int    `gorm:"default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// ----- Repository Interface -----

type UserRepository interface {
	CreateUser(u *User) error
	GetUserByID(id uint) (*User, error)
	UpdateUser(u *User) error
	DeleteUser(id uint) error
	GetUserByNickname(nickname string) (*User, error)

	// NOTE: Keep the interface as it was, returning []*UserWithLove
	TopLoveUsers(limit int) ([]*UserWithLove, error)
}

// ----- Implementation -----

type userRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository { return &userRepository{db: db} }

func (r *userRepository) CreateUser(u *User) error { return r.db.Create(u).Error }
func (r *userRepository) GetUserByID(id uint) (*User, error) {
	var u User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
func (r *userRepository) UpdateUser(u *User) error { return r.db.Save(u).Error }
func (r *userRepository) DeleteUser(id uint) error { return r.db.Delete(&User{}, id).Error }
func (r *userRepository) GetUserByNickname(nickname string) (*User, error) {
	var u User
	if err := r.db.Where("nickname = ?", nickname).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepository) TopLoveUsers(limit int) ([]*UserWithLove, error) {
	var users []*UserWithLove
	// Adjust "love_score" to match your actual column name.
	// If using embedded LoveMeter with prefix love_, and field Score, it might be "love_score".
	if err := r.db.Order("love_score DESC").Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

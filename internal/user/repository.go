package user

import "gorm.io/gorm"

type UserRepository interface {
	TopLoveMeter(limit int) ([]*User, error)
}

type userRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) TopLoveMeter(limit int) ([]*User, error) {
	var users []*User
	if err := r.db.Order("love_score DESC").Limit(limit).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

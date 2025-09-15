//go:generate mockgen -destination=mocks/mock_player_repository.go -package=mocks github.com/MyelinBots/catbot-go/internal/db/repositories/cat_player CatPlayerRepository
package cat_player

import (
	"context"
	"errors"

	"github.com/MyelinBots/catbot-go/internal/db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CatPlayerRepository interface {
	GetPlayerByID(id string) (*CatPlayer, error)
	GetAllPlayers(ctx context.Context, network string, channel string) ([]*CatPlayer, error)
	UpsertPlayer(ctx context.Context, player *CatPlayer) error
	GetPlayerByName(name string, network string, channel string) (*CatPlayer, error)
}

type CatPlayerRepositoryImpl struct {
	db *db.DB // wrapper with .DB *gorm.DB
}

func NewPlayerRepository(db *db.DB) CatPlayerRepository {
	return &CatPlayerRepositoryImpl{db: db}
}

func (r *CatPlayerRepositoryImpl) GetPlayerByID(id string) (*CatPlayer, error) {
	var p CatPlayer
	err := r.db.DB.Where("id = ?", id).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *CatPlayerRepositoryImpl) GetAllPlayers(ctx context.Context, network string, channel string) ([]*CatPlayer, error) {
	var players []*CatPlayer
	err := r.db.DB.WithContext(ctx).
		Where("network = ? AND channel = ?", network, channel).
		Find(&players).Error
	if err != nil {
		return nil, err
	}
	return players, nil
}

func (r *CatPlayerRepositoryImpl) GetPlayerByName(name string, network string, channel string) (*CatPlayer, error) {
	var player CatPlayer
	err := r.db.DB.
		Where("name = ? AND network = ? AND channel = ?", name, network, channel).
		First(&player).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &player, nil
}

func (r *CatPlayerRepositoryImpl) UpsertPlayer(ctx context.Context, player *CatPlayer) error {
	// Try to find an existing player by (name, channel, network)
	var existing CatPlayer
	err := r.db.DB.WithContext(ctx).
		Where("name = ? AND channel = ? AND network = ?", player.Name, player.Channel, player.Network).
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new
			if player.ID == "" {
				player.ID = uuid.NewString()
			}
			return r.db.DB.WithContext(ctx).Create(player).Error
		}
		// Some other DB error
		return err
	}

	// Update existing: reuse the same primary key and save the incoming struct.
	player.ID = existing.ID
	return r.db.DB.WithContext(ctx).Save(player).Error
}

package cat_player

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MyelinBots/catbot-go/internal/db"
	"gorm.io/gorm"
)

type CatPlayer struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string    `gorm:"column:name;type:text;not null;index:idx_player_scope,priority:1"`
	Network   string    `gorm:"column:network;type:text;not null;index:idx_player_scope,priority:2"`
	Channel   string    `gorm:"column:channel;type:text;not null;index:idx_player_scope,priority:3"`
	LoveMeter int       `gorm:"column:love_meter;type:int;not null;default:0"`
	Count     int       `gorm:"column:count;type:int;not null;default:0"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

type CatPlayerRepository interface {
	GetPlayerByID(id string) (*CatPlayer, error)
	GetAllPlayers(ctx context.Context, network, channel string) ([]*CatPlayer, error)
	UpsertPlayer(ctx context.Context, player *CatPlayer) error
	GetPlayerByName(name, network, channel string) (*CatPlayer, error)
	TopLoveMeter(ctx context.Context, network, channel string, limit int) ([]*CatPlayer, error)
}

type CatPlayerRepositoryImpl struct {
	db *db.DB // wrapper holding .DB *gorm.DB
}

func NewPlayerRepository(database *db.DB) CatPlayerRepository {
	return &CatPlayerRepositoryImpl{db: database}
}

func (r *CatPlayerRepositoryImpl) GetPlayerByID(id string) (*CatPlayer, error) {
	var p CatPlayer
	if err := r.db.DB.Where("id = ?", id).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *CatPlayerRepositoryImpl) GetAllPlayers(ctx context.Context, network, channel string) ([]*CatPlayer, error) {
	var players []*CatPlayer
	if err := r.db.DB.WithContext(ctx).
		Where("network = ? AND channel = ?", strings.ToLower(network), strings.ToLower(channel)).
		Find(&players).Error; err != nil {
		return nil, err
	}
	return players, nil
}

func (r *CatPlayerRepositoryImpl) GetPlayerByName(name, network, channel string) (*CatPlayer, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	network = strings.ToLower(network)
	channel = strings.ToLower(channel)

	var p CatPlayer
	if err := r.db.DB.
		Where("name = ? AND network = ? AND channel = ?", name, network, channel).
		First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *CatPlayerRepositoryImpl) UpsertPlayer(ctx context.Context, player *CatPlayer) error {
	// normalize for consistent uniqueness
	player.Name = strings.ToLower(strings.TrimSpace(player.Name))
	player.Network = strings.ToLower(player.Network)
	player.Channel = strings.ToLower(player.Channel)

	var existing CatPlayer
	err := r.db.DB.WithContext(ctx).
		Where("name = ? AND channel = ? AND network = ?", player.Name, player.Channel, player.Network).
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If you don't use Postgres/pgcrypto UUIDs, set: player.ID = uuid.NewString()
			return r.db.DB.WithContext(ctx).Create(player).Error
		}
		return err
	}

	// keep same primary key; update values
	player.ID = existing.ID
	return r.db.DB.WithContext(ctx).Save(player).Error
}

func (r *CatPlayerRepositoryImpl) TopLoveMeter(ctx context.Context, network, channel string, limit int) ([]*CatPlayer, error) {
	network = strings.ToLower(network)
	channel = strings.ToLower(channel)

	var players []*CatPlayer
	if err := r.db.DB.WithContext(ctx).
		Where("network = ? AND channel = ?", network, channel).
		Order("love_meter DESC").
		Limit(limit).
		Find(&players).Error; err != nil {
		return nil, err
	}
	return players, nil
}

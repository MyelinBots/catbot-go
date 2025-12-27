package cat_player

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/MyelinBots/catbot-go/internal/db"
	"gorm.io/gorm"
)

/*
Model
*/

type CatPlayer struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`

	Name    string `gorm:"column:name;type:text;not null;index:idx_player_scope,priority:1"`
	Network string `gorm:"column:network;type:text;not null;index:idx_player_scope,priority:2"`
	Channel string `gorm:"column:channel;type:text;not null;index:idx_player_scope,priority:3"`

	LoveMeter int `gorm:"column:love_meter;type:int;not null;default:0"`
	Count     int `gorm:"column:count;type:int;not null;default:0"`

	LastInteractedAt *time.Time `gorm:"column:last_interacted_at;index"`
	LastDecayAt      *time.Time `gorm:"column:last_decay_at;index"`

	PerfectDropWarned bool `gorm:"column:perfect_drop_warned;not null;default:false"`
}

/*
Repository interface
*/

type CatPlayerRepository interface {
	GetPlayerByID(ctx context.Context, id string) (*CatPlayer, error)
	GetPlayerByName(ctx context.Context, name, network, channel string) (*CatPlayer, error)
	GetAllPlayers(ctx context.Context, network, channel string) ([]*CatPlayer, error)

	UpsertPlayer(ctx context.Context, player *CatPlayer) error

	TopLoveMeter(ctx context.Context, network, channel string, limit int) ([]*CatPlayer, error)

	// ✅ helpers สำหรับ daily decay / analytics
	TouchInteraction(ctx context.Context, name, network, channel string, t time.Time) error
	SetDecayAt(ctx context.Context, name, network, channel string, t time.Time) error
	ListPlayersAtOrAbove(ctx context.Context, network, channel string, minLove int) ([]*CatPlayer, error)

	SetPerfectDropWarned(ctx context.Context, name, network, channel string, warned bool) error
}

/*
Repository impl
*/

type CatPlayerRepositoryImpl struct {
	db *db.DB // wrapper holding .DB *gorm.DB
}

func NewPlayerRepository(database *db.DB) CatPlayerRepository {
	return &CatPlayerRepositoryImpl{db: database}
}

/*
Normalization helpers
*/

func norm(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func normScope(network, channel string) (string, string) {
	return norm(network), norm(channel)
}

/*
CRUD
*/

func (r *CatPlayerRepositoryImpl) GetPlayerByID(ctx context.Context, id string) (*CatPlayer, error) {
	var p CatPlayer
	err := r.db.DB.WithContext(ctx).Where("id = ?", id).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *CatPlayerRepositoryImpl) GetAllPlayers(ctx context.Context, network, channel string) ([]*CatPlayer, error) {
	network, channel = normScope(network, channel)

	var players []*CatPlayer
	if err := r.db.DB.WithContext(ctx).
		Where("network = ? AND channel = ?", network, channel).
		Find(&players).Error; err != nil {
		return nil, err
	}
	return players, nil
}

func (r *CatPlayerRepositoryImpl) GetPlayerByName(ctx context.Context, name, network, channel string) (*CatPlayer, error) {
	name = norm(name)
	network, channel = normScope(network, channel)

	var p CatPlayer
	err := r.db.DB.WithContext(ctx).
		Where("name = ? AND network = ? AND channel = ?", name, network, channel).
		First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// Upsert by (name, network, channel)
func (r *CatPlayerRepositoryImpl) UpsertPlayer(ctx context.Context, player *CatPlayer) error {
	// normalize for consistent uniqueness
	player.Name = norm(player.Name)
	player.Network, player.Channel = normScope(player.Network, player.Channel)

	var existing CatPlayer
	err := r.db.DB.WithContext(ctx).
		Where("name = ? AND network = ? AND channel = ?", player.Name, player.Network, player.Channel).
		First(&existing).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.db.DB.WithContext(ctx).Create(player).Error
		}
		return err
	}

	// keep same primary key; update values
	player.ID = existing.ID
	return r.db.DB.WithContext(ctx).Save(player).Error
}

/*
Leaderboard
*/

func (r *CatPlayerRepositoryImpl) TopLoveMeter(ctx context.Context, network, channel string, limit int) ([]*CatPlayer, error) {
	network, channel = normScope(network, channel)
	if limit <= 0 {
		limit = 5
	}

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

/*
Daily-decay helpers
*/

func (r *CatPlayerRepositoryImpl) TouchInteraction(ctx context.Context, name, network, channel string, t time.Time) error {
	name = norm(name)
	network, channel = normScope(network, channel)

	return r.db.DB.WithContext(ctx).
		Model(&CatPlayer{}).
		Where("name = ? AND network = ? AND channel = ?", name, network, channel).
		Updates(map[string]any{"last_interacted_at": &t}).Error
}

func (r *CatPlayerRepositoryImpl) SetDecayAt(ctx context.Context, name, network, channel string, t time.Time) error {
	name = norm(name)
	network, channel = normScope(network, channel)

	return r.db.DB.WithContext(ctx).
		Model(&CatPlayer{}).
		Where("name = ? AND network = ? AND channel = ?", name, network, channel).
		Updates(map[string]any{"last_decay_at": &t}).Error
}

func (r *CatPlayerRepositoryImpl) ListPlayersAtOrAbove(ctx context.Context, network, channel string, minLove int) ([]*CatPlayer, error) {
	network, channel = normScope(network, channel)

	var players []*CatPlayer
	if err := r.db.DB.WithContext(ctx).
		Where("network = ? AND channel = ? AND love_meter >= ?", network, channel, minLove).
		Find(&players).Error; err != nil {
		return nil, err
	}
	return players, nil
}

func (r *CatPlayerRepositoryImpl) SetPerfectDropWarned(ctx context.Context, name, network, channel string, warned bool) error {
	name = norm(name)
	network, channel = normScope(network, channel)

	return r.db.DB.WithContext(ctx).
		Model(&CatPlayer{}).
		Where("name = ? AND network = ? AND channel = ?", name, network, channel).
		Updates(map[string]any{"perfect_drop_warned": warned}).Error
}

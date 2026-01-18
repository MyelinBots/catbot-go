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
MODEL
*/

type CatPlayer struct {
	ID        uint      `gorm:"column:id;primaryKey;autoIncrement" json:"id"` // ✅ MySQL friendly
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`

	Name    string `gorm:"column:name;type:varchar(100);not null;index:idx_player_scope,priority:1"`
	Network string `gorm:"column:network;type:varchar(100);not null;index:idx_player_scope,priority:2"`
	Channel string `gorm:"column:channel;type:varchar(100);not null;index:idx_player_scope,priority:3"`

	LoveMeter int `gorm:"column:love_meter;type:int;not null;default:0"`
	Count     int `gorm:"column:count;type:int;not null;default:0"`

	LastInteractedAt *time.Time `gorm:"column:last_interacted_at;index"`
	LastDecayAt      *time.Time `gorm:"column:last_decay_at;index"`

	PerfectDropWarned bool `gorm:"column:perfect_drop_warned;not null;default:false"`

	// ✅ Bond system (independent from lovemeter except gate love==100)
	BondPoints        int        `gorm:"column:bond_points;type:int;not null;default:0"`
	BondPointStreak   int        `gorm:"column:bond_point_streak;type:int;not null;default:0"`
	HighestBondStreak int        `gorm:"column:highest_bond_streak;type:int;not null;default:0"`
	LastBondPointsAt  *time.Time `gorm:"column:last_bond_points_at;index"`

	// bitmask gifts
	GiftsUnlocked int `gorm:"column:gifts_unlocked;type:int;not null;default:0"`
}

/*
REPOSITORY INTERFACE
*/

type CatPlayerRepository interface {
	GetPlayerByID(ctx context.Context, id uint) (*CatPlayer, error)
	GetPlayerByName(ctx context.Context, name, network, channel string) (*CatPlayer, error)
	GetAllPlayers(ctx context.Context, network, channel string) ([]*CatPlayer, error)

	UpsertPlayer(ctx context.Context, player *CatPlayer) error
	TopLoveMeter(ctx context.Context, network, channel string, limit int) ([]*CatPlayer, error)

	// daily decay helpers
	TouchInteraction(ctx context.Context, name, network, channel string, t time.Time) error
	SetDecayAt(ctx context.Context, name, network, channel string, t time.Time) error
	ListPlayersAtOrAbove(ctx context.Context, network, channel string, minLove int) ([]*CatPlayer, error)
	SetPerfectDropWarned(ctx context.Context, name, network, channel string, warned bool) error

	// bond helpers
	AddBondPoints(ctx context.Context, name, network, channel string, delta int) error
	SetBondPointsAt(ctx context.Context, name, network, channel string, t time.Time) error
	SetBondPointStreak(ctx context.Context, name, network, channel string, streak int) error
	SetHighestBondStreak(ctx context.Context, name, network, channel string, streak int) error

	// gifts (bitmask)
	AddGiftsUnlocked(ctx context.Context, name, network, channel string, giftMask int) error
	SetGiftsUnlocked(ctx context.Context, name, network, channel string, giftsUnlocked int) error
}

/*
REPOSITORY IMPL
*/

type CatPlayerRepositoryImpl struct {
	db *db.DB // wrapper holding .DB *gorm.DB
}

func NewPlayerRepository(database *db.DB) CatPlayerRepository {
	return &CatPlayerRepositoryImpl{db: database}
}

/*
NORMALIZATION
*/

func norm(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func normScope(network, channel string) (string, string) {
	return norm(network), norm(channel)
}

/*
CRUD
*/

func (r *CatPlayerRepositoryImpl) GetPlayerByID(ctx context.Context, id uint) (*CatPlayer, error) {
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

	// keep same primary key
	player.ID = existing.ID
	return r.db.DB.WithContext(ctx).Save(player).Error
}

/*
LEADERBOARD
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
DAILY DECAY HELPERS
*/

func (r *CatPlayerRepositoryImpl) TouchInteraction(ctx context.Context, name, network, channel string, t time.Time) error {
	name = norm(name)
	network, channel = normScope(network, channel)

	return r.db.DB.WithContext(ctx).
		Model(&CatPlayer{}).
		Where("name = ? AND network = ? AND channel = ?", name, network, channel).
		Update("last_interacted_at", &t).Error
}

func (r *CatPlayerRepositoryImpl) SetDecayAt(ctx context.Context, name, network, channel string, t time.Time) error {
	name = norm(name)
	network, channel = normScope(network, channel)

	return r.db.DB.WithContext(ctx).
		Model(&CatPlayer{}).
		Where("name = ? AND network = ? AND channel = ?", name, network, channel).
		Update("last_decay_at", &t).Error
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
		Update("perfect_drop_warned", warned).Error
}

/*
BOND HELPERS
*/

func (r *CatPlayerRepositoryImpl) AddBondPoints(ctx context.Context, name, network, channel string, delta int) error {
	name = norm(name)
	network, channel = normScope(network, channel)

	return r.db.DB.WithContext(ctx).
		Model(&CatPlayer{}).
		Where("name = ? AND network = ? AND channel = ?", name, network, channel).
		UpdateColumn("bond_points", gorm.Expr("bond_points + ?", delta)).Error
}

func (r *CatPlayerRepositoryImpl) SetBondPointsAt(ctx context.Context, name, network, channel string, t time.Time) error {
	name = norm(name)
	network, channel = normScope(network, channel)

	return r.db.DB.WithContext(ctx).
		Model(&CatPlayer{}).
		Where("name = ? AND network = ? AND channel = ?", name, network, channel).
		Update("last_bond_points_at", &t).Error
}

func (r *CatPlayerRepositoryImpl) SetBondPointStreak(ctx context.Context, name, network, channel string, streak int) error {
	name = norm(name)
	network, channel = normScope(network, channel)

	return r.db.DB.WithContext(ctx).
		Model(&CatPlayer{}).
		Where("name = ? AND network = ? AND channel = ?", name, network, channel).
		Update("bond_point_streak", streak).Error
}

func (r *CatPlayerRepositoryImpl) SetHighestBondStreak(ctx context.Context, name, network, channel string, streak int) error {
	name = norm(name)
	network, channel = normScope(network, channel)

	return r.db.DB.WithContext(ctx).
		Model(&CatPlayer{}).
		Where("name = ? AND network = ? AND channel = ?", name, network, channel).
		Update("highest_bond_streak", streak).Error
}

/*
GIFTS (bitmask)
MySQL-safe: use SQL bitwise OR
*/

func (r *CatPlayerRepositoryImpl) AddGiftsUnlocked(ctx context.Context, name, network, channel string, giftMask int) error {
	name = norm(name)
	network, channel = normScope(network, channel)

	return r.db.DB.WithContext(ctx).Exec(
		`UPDATE cat_player
		 SET gifts_unlocked = (gifts_unlocked | ?)
		 WHERE name = ? AND network = ? AND channel = ?`,
		giftMask, name, network, channel,
	).Error
}

func (r *CatPlayerRepositoryImpl) SetGiftsUnlocked(ctx context.Context, name, network, channel string, giftsUnlocked int) error {
	name = norm(name)
	network, channel = normScope(network, channel)

	return r.db.DB.WithContext(ctx).
		Model(&CatPlayer{}).
		Where("name = ? AND network = ? AND channel = ?", name, network, channel).
		Update("gifts_unlocked", giftsUnlocked).Error
}

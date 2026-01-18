
CREATE TABLE cat_player (
    nickname VARCHAR(100) UNIQUE NOT NULL,
    lovemeter INT NOT NULL DEFAULT 0,
    PRIMARY KEY (nickname)
);

ALTER TABLE cat_player
  ADD CONSTRAINT cat_player_name_network_channel_unique
  UNIQUE (name, network, channel);

select id, name, network, channel, love_meter
from cat_player
where lower(network)=lower('YOUR_NETWORK') and lower(channel)=lower('#darkworld')
order by love_meter desc
limit 5;

ALTER TABLE cat_player
ADD COLUMN loyalty_streak INT NOT NULL DEFAULT 0,
ADD COLUMN bond_points INT NOT NULL DEFAULT 0,
ADD COLUMN last_interacted_at DATETIME NULL,
ADD COLUMN last_bond_points_at DATETIME NULL;
ADD COLUMN last_decay_at DATETIME NULL,
ADD COLUMN perfect_drop_warned TINYINT(1) NOT NULL DEFAULT 0;

-- =========================================
-- MySQL Migration: BondPoints + Gifts + Streak
-- =========================================

-- --- BondPoints system ---
ALTER TABLE cat_player ADD COLUMN bond_points INT NOT NULL DEFAULT 0;
ALTER TABLE cat_player ADD COLUMN bond_point_streak INT NOT NULL DEFAULT 0;
ALTER TABLE cat_player ADD COLUMN highest_bond_streak INT NOT NULL DEFAULT 0;
ALTER TABLE cat_player ADD COLUMN last_bond_points_at DATETIME NULL;

-- --- Interaction / decay support ---
ALTER TABLE cat_player ADD COLUMN last_interacted_at DATETIME NULL;
ALTER TABLE cat_player ADD COLUMN last_decay_at DATETIME NULL;

-- --- Gifts / Titles unlock tracking (bitmask) ---
-- gifts_unlocked bits:
-- 1 = 7-day gift
-- 2 = 14-day gift
-- 4 = 30-day gift
ALTER TABLE cat_player ADD COLUMN gifts_unlocked INT NOT NULL DEFAULT 0;

-- --- Optional warning flag (if you use it)
ALTER TABLE cat_player ADD COLUMN perfect_drop_warned TINYINT(1) NOT NULL DEFAULT 0;

-- Add bond points system columns
ALTER TABLE cat_player
    ADD COLUMN bond_points INT NOT NULL DEFAULT 0,
    ADD COLUMN bond_point_streak INT NOT NULL DEFAULT 0,
    ADD COLUMN highest_bond_streak INT NOT NULL DEFAULT 0,
    ADD COLUMN last_bond_points_at TIMESTAMP NULL,
    ADD COLUMN gifts_unlocked INT NOT NULL DEFAULT 0;
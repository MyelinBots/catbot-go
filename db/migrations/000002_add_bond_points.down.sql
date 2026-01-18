-- Remove bond points system columns
ALTER TABLE cat_player
    DROP COLUMN bond_points,
    DROP COLUMN bond_point_streak,
    DROP COLUMN highest_bond_streak,
    DROP COLUMN last_bond_points_at,
    DROP COLUMN last_interacted_at,
    DROP COLUMN last_decay_at,
    DROP COLUMN gifts_unlocked,
    DROP COLUMN perfect_drop_warned;

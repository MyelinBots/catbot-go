CREATE TABLE cat_player (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    name TEXT NOT NULL,
    network TEXT NOT NULL,
    channel TEXT NOT NULL,
    love_meter INT NOT NULL DEFAULT 0,
    count INT NOT NULL DEFAULT 0,
    last_interacted_at TIMESTAMP NULL,
    last_decay_at TIMESTAMP NULL,
    perfect_drop_warned BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_player_scope ON cat_player (name, network, channel);
CREATE UNIQUE INDEX idx_player_unique ON cat_player (name, network, channel);

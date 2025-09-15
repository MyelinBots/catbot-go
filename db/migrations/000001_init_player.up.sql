
CREATE TABLE cat_player (
    nickname VARCHAR(100) UNIQUE NOT NULL,
    lovemeter INT NOT NULL DEFAULT 0,
    PRIMARY KEY (nickname)
);
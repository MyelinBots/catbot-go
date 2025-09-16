
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

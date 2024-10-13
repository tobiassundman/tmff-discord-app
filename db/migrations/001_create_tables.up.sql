-- players table contains all players that participate in the league
CREATE TABLE IF NOT EXISTS players (
    bga_id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- seasons table contains all seasons that have been played
CREATE TABLE IF NOT EXISTS seasons (
    name TEXT PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- season_participants table contains all players that have participated in a season and their season elo
CREATE TABLE IF NOT EXISTS season_participants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    season_name TEXT NOT NULL,
    player_id TEXT NOT NULL,
    elo INTEGER NOT NULL,
    games_played INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE(season_name, player_id),
    FOREIGN KEY(season_name) REFERENCES seasons(name),
    FOREIGN KEY(player_id) REFERENCES players(bga_id)
);

-- games table contains all games that have been played in a specific season
CREATE TABLE IF NOT EXISTS games (
    bga_id TEXT PRIMARY KEY,
    season_name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY(season_name) REFERENCES seasons(name)
);

-- game_participants table contains all players that have participated in a game and their score and elo change
CREATE TABLE IF NOT EXISTS game_participants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id TEXT NOT NULL,
    player_id TEXT NOT NULL,
    score INTEGER NOT NULL,
    elo_change INTEGER NOT NULL,
    elo_before INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE(game_id, player_id),
    FOREIGN KEY(game_id) REFERENCES games(bga_id),
    FOREIGN KEY(player_id) REFERENCES players(bga_id)
);

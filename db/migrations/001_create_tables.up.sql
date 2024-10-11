CREATE TABLE IF NOT EXISTS players (
    bga_id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS seasons (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS season_participants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    season_id INTEGER NOT NULL,
    FOREIGN KEY(season_id) REFERENCES seasons(id),
    player_id TEXT NOT NULL,
    FOREIGN KEY(player_id) REFERENCES players(bga_id),
    elo INTEGER NOT NULL,
    games_played INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE(season_id, player_id)
);

CREATE TABLE IF NOT EXISTS games (
    bga_id TEXT PRIMARY KEY,
    season_id INTEGER NOT NULL,
    FOREIGN KEY(season_id) REFERENCES seasons(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS game_participants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_id TEXT NOT NULL,
    FOREIGN KEY(game_id) REFERENCES games(bga_id),
    player_id TEXT NOT NULL,
    FOREIGN KEY(player_id) REFERENCES players(bga_id),
    score INTEGER NOT NULL,
    elo_change INTEGER NOT NULL,
    elo_before INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE(game_id, player_id)
);
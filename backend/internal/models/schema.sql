DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS live_matches;
DROP TABLE IF EXISTS past_matches;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS user_ratings;

CREATE TABLE sessions (
	token TEXT PRIMARY KEY,
	data BLOB NOT NULL,
	expiry REAL NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions(expiry);

CREATE TABLE live_matches (
    match_id INTEGER PRIMARY KEY AUTOINCREMENT, 
    white_player_id INTEGER NOT NULL, 
    black_player_id INTEGER NOT NULL,
    last_move_piece INTEGER,
    last_move_move INTEGER,
    current_fen TEXT DEFAULT 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1' NOT NULL,
    time_format_in_milliseconds INTEGER NOT NULL,
    increment_in_milliseconds INTEGER NOT NULL,
    white_player_time_remaining_in_milliseconds INTEGER NOT NULL,
    black_player_time_remaining_in_milliseconds INTEGER NOT NULL,
    game_history_json_string BLOB NOT NULL,
    unix_ms_time_of_last_move INTEGER NOT NULL,
    average_elo REAL NOT NULL,
    white_player_elo REAL NOT NULL,
    black_player_elo REAL NOT NULL,
    match_start_time INTEGER NOT NULL
);

CREATE TABLE past_matches (
    match_id INTEGER PRIMARY KEY NOT NULL, 
    white_player_id INTEGER NOT NULL, 
    black_player_id INTEGER NOT NULL,
    last_move_piece INTEGER NOT NULL,
    last_move_move INTEGER NOT NULL,
    final_fen TEXT NOT NULL,
    time_format_in_milliseconds INTEGER NOT NULL,
    increment_in_milliseconds INTEGER NOT NULL,
    game_history_json_string BLOB NOT NULL,
    result INTEGER NOT NULL,
    result_reason INTEGER NOT NULL,
    white_player_elo REAL NOT NULL,
    black_player_elo REAL NOT NULL,
    white_player_elo_gain REAL NOT NULL,
    black_player_elo_gain REAL NOT NULL,
    average_elo REAL NOT NULL,
    match_start_time INTEGER NOT NULL,
    match_end_time INTEGER NOT NULL
);

CREATE TABLE users (
    player_id INTEGER PRIMARY KEY NOT NULL,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    email TEXT,
    join_date INTEGER DEFAULT (strftime('%s', 'now')),
    last_seen INTEGER DEFAULT (strftime('%s', 'now'))
);

CREATE UNIQUE INDEX users_username_idx ON users (username);

CREATE TABLE user_ratings (
    player_id INTEGER PRIMARY KEY NOT NULL,
    username TEXT UNIQUE NOT NULL,
    bullet_rating INTEGER DEFAULT 1500,
    blitz_rating INTEGER DEFAULT 1500,
    rapid_rating INTEGER DEFAULT 1500,
    classical_rating INTEGER DEFAULT 1500
);

CREATE UNIQUE INDEX user_ratings_username_idx ON user_ratings (username);
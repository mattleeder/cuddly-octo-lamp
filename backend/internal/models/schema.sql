DROP TABLE IF EXISTS live_matches;
DROP TABLE IF EXISTS past_matches;
DROP TABLE IF EXISTS users;

CREATE TABLE live_matches (
    match_id INTEGER PRIMARY KEY NOT NULL, 
    white_player_id INTEGER NOT NULL, 
    black_player_id INTEGER NOT NULL,
    last_move_piece INTEGER,
    last_move_move INTEGER,
    current_fen TEXT DEFAULT 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1',
    time_format_in_milliseconds INTEGER NOT NULL,
    increment_in_milliseconds INTEGER NOT NULL,
    white_player_time_remaining_in_milliseconds INTEGER NOT NULL,
    black_player_time_remaining_in_milliseconds INTEGER NOT NULL,
    game_history_json_string BLOB NOT NULL,
    unix_ms_time_of_last_move INTEGER NOT NULL,
    average_elo REAL NOT NULL
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
    white_player_points REAL NOT NULL,
    black_player_points REAL NOT NULL,
    average_elo REAL NOT NULL
);

CREATE TABLE users (
    player_id INTEGER PRIMARY KEY NOT NULL,
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    email TEXT,
    join_date INTEGER DEFAULT unixepoch('now'),
    last_seen INTEGER DEFAULT unixepoch('now')
);

CREATE TABLE user_ratings (
    player_id INTEGER PRIMARY KEY NOT NULL,
    username TEXT NOT NULL,
    bullet_rating INTEGER NOT NULL DEFAULT 1500,
    blitz_rating INTEGER NOT NULL DEFAULT 1500,
    rapid_rating INTEGER NOT NULL DEFAULT 1500,
    classical_rating INTEGER NOT NULL DEFAULT 1500
);
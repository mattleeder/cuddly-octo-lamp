DROP TABLE IF EXISTS live_matches;
DROP TABLE IF EXISTS past_matches;

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
    unix_ms_time_of_last_move INTEGER NOT NULL
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
    black_player_points REAL NOT NULL
)
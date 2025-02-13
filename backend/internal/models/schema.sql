DROP TABLE IF EXISTS live_matches;
DROP TABLE IF EXISTS matchmaking_queue;

CREATE TABLE live_matches (
    match_id INTEGER PRIMARY KEY NOT NULL, 
    white_player_id INTEGER NOT NULL, 
    black_player_id INTEGER NOT NULL,
    last_move_piece INTEGER,
    last_move_move INTEGER,
    current_fen TEXT DEFAULT "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
);

CREATE TABLE matchmaking_queue (
    playerid INTEGER PRIMARY KEY NOT NULL
);
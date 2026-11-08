//! Texts shown by the client, in the course language (日本語).
//! `./mudschool lang <sv|en|ja>` replaces this file.

pub const CONNECTED_TO: &str = "接続先:";
pub const SERVER_CLOSED: &str = "サーバーが接続を切りました。";
pub const DISCONNECTED: &str = "*** サーバーとの接続が切れました。Esc で終了します。";
pub const NOT_LOGGED_IN: &str = " 未ログイン";
pub const NO_MAP: &str = "（マップなし）";
pub const TITLE_LOG: &str = " ログ ";
pub const TITLE_MAP: &str = " マップ ";
pub const TITLE_INPUT: &str = " コマンド ";
pub const ONLINE: &str = "人がオンライン";
pub const STARTER: &str = "MUD クライアントからこんにちは！このファイルはレッスン 3 で書きます。";
pub const TITLE_PASSWORD: &str = " パスワード（非表示） ";
pub const CONFIG_ERROR: &str = "設定にエラーがあります";
// The mouse (lesson 30): menu entries and what tiles are called.
pub const LOOK_AT: &str = "調べる";
pub const ATTACK: &str = "攻撃する";
pub const PICK_UP: &str = "拾う";
pub const WALK_HERE: &str = "ここへ歩く";

/// What a tile meaning is called, for the mouse pointer's tooltip.
pub fn tile_name(meaning: &str) -> &'static str {
    match meaning {
        "floor" => "床",
        "wall" => "壁",
        "grass" => "草",
        "tree" => "木",
        "water" => "水",
        "rock" => "岩",
        "door" => "扉",
        "stairs_up" => "上り階段",
        "stairs_down" => "下り階段",
        _ => "何か",
    }
}

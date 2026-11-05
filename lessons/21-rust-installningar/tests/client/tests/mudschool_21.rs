// This file belongs to the course tool and is overwritten by `mudschool check`.

mod mudschool_common;

use std::path::PathBuf;

use clap::Parser;
use ratatui::backend::TestBackend;
use ratatui::crossterm::event::{KeyCode, KeyEvent, KeyEventKind, KeyEventState, KeyModifiers};
use ratatui::Terminal;

use mud_client::app::{draw, App};
use mud_client::config::{Cli, Config, ConfigError, FileConfig, DEFAULT_SERVER};
use mud_client::messages;
use mud_client::protocol::{parse_line, ServerMsg, HELLO};
use mudschool_common::l;

fn key(code: KeyCode) -> KeyEvent {
    KeyEvent {
        code,
        modifiers: KeyModifiers::NONE,
        kind: KeyEventKind::Press,
        state: KeyEventState::NONE,
    }
}

fn type_line(app: &mut App, text: &str) -> Option<String> {
    for c in text.chars() {
        app.on_key(key(KeyCode::Char(c)));
    }
    app.on_key(key(KeyCode::Enter))
}

fn screen(app: &App) -> String {
    let mut terminal = Terminal::new(TestBackend::new(80, 16)).unwrap();
    terminal.draw(|f| draw(f, app)).unwrap();
    let buf = terminal.backend().buffer();
    let mut out = String::new();
    for y in 0..buf.area.height {
        for x in 0..buf.area.width {
            out.push_str(buf[(x, y)].symbol());
        }
        out.push('\n');
    }
    out
}

#[test]
fn parses_prompt_frames() {
    assert_eq!(
        parse_line(r#"{"type":"prompt","prompt":"password"}"#),
        ServerMsg::Prompt {
            prompt: "password".into()
        },
        "{}",
        l(
            "en prompt-frame ska bli ServerMsg::Prompt",
            "a prompt frame should become ServerMsg::Prompt",
            "prompt フレームは ServerMsg::Prompt になるべきです"
        )
    );
}

#[test]
fn passwords_are_hidden() {
    let mut app = App::new();
    app.on_server(parse_line(r#"{"type":"prompt","prompt":"password"}"#));
    assert!(
        app.secret,
        "{}",
        l(
            "efter en lösenordsprompt ska app.secret vara true",
            "after a password prompt app.secret should be true",
            "パスワードの prompt の後は app.secret が true であるべきです"
        )
    );
    for c in "hemligt".chars() {
        app.on_key(key(KeyCode::Char(c)));
    }
    let shown = screen(&app);
    assert!(
        !shown.contains("hemligt") && shown.contains("*******"),
        "{}\n{shown}",
        l(
            "lösenordet ska visas som ******* – inte i klartext",
            "the password should be shown as ******* – not in plain text",
            "パスワードは平文ではなく ******* と表示されるべきです"
        )
    );
    // Wide characters (e.g. Japanese) take two cells, so compare without spaces.
    let squeeze = |s: &str| s.replace(' ', "");
    assert!(
        squeeze(&shown).contains(&squeeze(messages::TITLE_PASSWORD)),
        "{}\n{shown}",
        l(
            "rutans rubrik ska vara messages::TITLE_PASSWORD",
            "the box title should be messages::TITLE_PASSWORD",
            "枠の見出しは messages::TITLE_PASSWORD であるべきです"
        )
    );

    assert_eq!(app.on_key(key(KeyCode::Enter)), Some("hemligt".to_string()));
    assert!(!app.secret, "{}", l(
        "efter Enter ska inmatningen synas igen",
        "after Enter the input should be visible again",
        "Enter の後は入力が再び見えるべきです"
    ));
    let leaked = app.log.iter().any(|line| line.text.contains("hemligt"))
        || app.history.iter().any(|h| h.contains("hemligt"));
    assert!(!leaked, "{}", l(
        "lösenordet får inte hamna i loggen eller historiken",
        "the password must not end up in the log or the history",
        "パスワードがログや履歴に残ってはいけません"
    ));

    // Up must not bring anything secret back – and normal lines still work.
    assert_eq!(type_line(&mut app, "look"), Some("look".to_string()));
    assert_eq!(app.history, vec!["look".to_string()]);
}

#[test]
fn other_prompts_are_not_secret() {
    let mut app = App::new();
    app.on_server(parse_line(r#"{"type":"prompt","prompt":"new_password"}"#));
    assert!(app.secret);
    app.on_server(parse_line(r#"{"type":"prompt","prompt":"name"}"#));
    assert!(!app.secret, "{}", l(
        "bara password och new_password ska döljas",
        "only password and new_password should be hidden",
        "隠すのは password と new_password だけです"
    ));
}

#[test]
fn command_line() {
    let cli = Cli::try_parse_from(["mud-client", "--name", "Anna", "10.0.0.5:4000"]).unwrap();
    assert_eq!(cli.server.as_deref(), Some("10.0.0.5:4000"));
    assert_eq!(cli.name.as_deref(), Some("Anna"));
    assert_eq!(cli.config, None);

    let cli = Cli::try_parse_from(["mud-client", "--config", "other.toml"]).unwrap();
    assert_eq!(cli.config, Some(PathBuf::from("other.toml")));
    assert_eq!(cli.server, None);

    assert!(
        Cli::try_parse_from(["mud-client", "--colour", "pink"]).is_err(),
        "{}",
        l(
            "en okänd flagga ska ge ett fel",
            "an unknown flag should give an error",
            "未知のフラグはエラーになるべきです"
        )
    );
}

#[test]
fn config_file() {
    let file = FileConfig::parse("server = \"mud.example.com:4000\"\nname = \"Bertil\"\n").unwrap();
    assert_eq!(file.server.as_deref(), Some("mud.example.com:4000"));
    assert_eq!(file.name.as_deref(), Some("Bertil"));
    assert_eq!(FileConfig::parse("").unwrap(), FileConfig::default());
    assert!(
        FileConfig::parse("sever = \"oops\"").is_err(),
        "{}",
        l(
            "ett felstavat fält ska ge ett fel (deny_unknown_fields)",
            "a misspelt field should give an error (deny_unknown_fields)",
            "綴りを間違えたフィールドはエラーになるべきです（deny_unknown_fields）"
        )
    );
}

#[test]
fn loading_files() {
    let dir = std::env::temp_dir().join(format!("mudschool-21-{}", std::process::id()));
    std::fs::create_dir_all(&dir).unwrap();

    let missing = dir.join("missing.toml");
    assert_eq!(
        FileConfig::load(&missing).unwrap(),
        FileConfig::default(),
        "{}",
        l(
            "en fil som inte finns betyder inga inställningar – inget fel",
            "a file that doesn't exist means no settings – not an error",
            "存在しないファイルは「設定なし」を意味し、エラーではありません"
        )
    );

    let bad = dir.join("bad.toml");
    std::fs::write(&bad, "server = 4000\n").unwrap();
    match FileConfig::load(&bad) {
        Err(err @ ConfigError::Parse { .. }) => {
            let text = err.to_string();
            assert!(
                text.contains("bad.toml"),
                "{}: {text}",
                l(
                    "felmeddelandet ska nämna filen",
                    "the error message should mention the file",
                    "エラーメッセージにはファイル名が含まれるべきです"
                )
            );
            assert!(std::error::Error::source(&err).is_some(), "#[source]");
        }
        other => panic!("{}: {other:?}", l("väntade ConfigError::Parse", "expected ConfigError::Parse", "ConfigError::Parse を期待しました")),
    }

    let good = dir.join("good.toml");
    std::fs::write(&good, "name = \"Cecilia\"\n").unwrap();
    let cli = Cli {
        config: Some(good),
        ..Cli::default()
    };
    let config = Config::load(&cli).unwrap();
    assert_eq!(config.name.as_deref(), Some("Cecilia"));
    assert_eq!(config.server, DEFAULT_SERVER);
    std::fs::remove_dir_all(&dir).ok();
}

#[test]
fn command_line_wins() {
    let file = FileConfig {
        server: Some("file:1".into()),
        name: Some("FromFile".into()),
    };
    let cli = Cli {
        server: Some("cli:2".into()),
        ..Cli::default()
    };
    let config = Config::merge(&cli, &file);
    assert_eq!(
        config,
        Config {
            server: "cli:2".into(),
            name: Some("FromFile".into())
        },
        "{}",
        l(
            "kommandoraden går före filen, filen före standardvärdena",
            "the command line wins over the file, the file over the defaults",
            "コマンドラインはファイルより、ファイルは既定値より優先されます"
        )
    );
    let config = Config::merge(&Cli::default(), &FileConfig::default());
    assert_eq!(config.server, DEFAULT_SERVER);
    assert_eq!(config.name, None);
}

#[test]
fn logs_in_automatically() {
    let config = Config {
        server: DEFAULT_SERVER.into(),
        name: Some("Anna".into()),
    };
    assert_eq!(config.startup_lines(), vec![HELLO.to_string(), "Anna".to_string()]);
    let config = Config {
        name: None,
        ..config
    };
    assert_eq!(config.startup_lines(), vec![HELLO.to_string()]);
}

//! Settings: command-line arguments, a config file and defaults.
//!
//! Precedence: what you type on the command line wins over the file, and the
//! file wins over the defaults.

use std::fs;
use std::io;
use std::path::{Path, PathBuf};

use clap::Parser;
use serde::Deserialize;

/// The server the client connects to if nothing else is said.
pub const DEFAULT_SERVER: &str = "127.0.0.1:4000";

/// The config file that is read if --config isn't given.
pub const DEFAULT_CONFIG: &str = "mudclient.toml";

/// The command line. clap generates the parser – and --help – from the struct.
#[derive(Debug, Clone, Default, PartialEq, Parser)]
#[command(name = "mud-client", version, about = "The MUD School client")]
pub struct Cli {
    /// The server to connect to, e.g. 127.0.0.1:4000
    pub server: Option<String>,
    /// Log in with this name automatically
    #[arg(long)]
    pub name: Option<String>,
    /// The config file (default: mudclient.toml)
    #[arg(long)]
    pub config: Option<PathBuf>,
}

/// What can be written in the config file. Every field is optional.
#[derive(Debug, Clone, Default, PartialEq, Deserialize)]
#[serde(default, deny_unknown_fields)]
pub struct FileConfig {
    pub server: Option<String>,
    pub name: Option<String>,
}

/// The settings the client runs with, after everything has been merged.
#[derive(Debug, Clone, PartialEq)]
pub struct Config {
    pub server: String,
    pub name: Option<String>,
}

/// Everything that can go wrong with the settings. thiserror writes the
/// Display and Error implementations from the #[error] attributes.
#[derive(Debug, thiserror::Error)]
pub enum ConfigError {
    #[error("{}: {source}", path.display())]
    Read {
        path: PathBuf,
        #[source]
        source: io::Error,
    },
    #[error("{}: {source}", path.display())]
    Parse {
        path: PathBuf,
        #[source]
        source: toml::de::Error,
    },
}

impl FileConfig {
    /// Parses the contents of a config file.
    pub fn parse(text: &str) -> Result<FileConfig, toml::de::Error> {
        toml::from_str(text)
    }

    /// Reads a config file. A file that doesn't exist is not an error: it
    /// means "no settings".
    pub fn load(path: &Path) -> Result<FileConfig, ConfigError> {
        let text = match fs::read_to_string(path) {
            Ok(text) => text,
            Err(e) if e.kind() == io::ErrorKind::NotFound => return Ok(FileConfig::default()),
            Err(source) => {
                return Err(ConfigError::Read {
                    path: path.to_path_buf(),
                    source,
                })
            }
        };
        FileConfig::parse(&text).map_err(|source| ConfigError::Parse {
            path: path.to_path_buf(),
            source,
        })
    }
}

impl Config {
    /// Merges the command line, the file and the defaults.
    pub fn merge(cli: &Cli, file: &FileConfig) -> Config {
        Config {
            server: cli
                .server
                .clone()
                .or_else(|| file.server.clone())
                .unwrap_or_else(|| DEFAULT_SERVER.to_string()),
            name: cli.name.clone().or_else(|| file.name.clone()),
        }
    }

    /// Reads the file the command line points at (or the default) and merges.
    pub fn load(cli: &Cli) -> Result<Config, ConfigError> {
        let path = cli
            .config
            .clone()
            .unwrap_or_else(|| PathBuf::from(DEFAULT_CONFIG));
        let file = FileConfig::load(&path)?;
        Ok(Config::merge(cli, &file))
    }

    /// The lines the client sends as soon as it has connected: the handshake,
    /// and the name if it is known.
    pub fn startup_lines(&self) -> Vec<String> {
        let mut lines = vec![crate::protocol::HELLO.to_string()];
        lines.extend(self.name.clone());
        lines
    }
}

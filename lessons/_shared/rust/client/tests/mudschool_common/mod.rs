// This file belongs to the course tool and is overwritten by `mudschool check`.
#![allow(dead_code)]

/// Picks the test message in the course language ($MUDSCHOOL_LANG).
pub fn l(sv: &'static str, en: &'static str, ja: &'static str) -> &'static str {
    match std::env::var("MUDSCHOOL_LANG").as_deref() {
        Ok("sv") => sv,
        Ok("ja") => ja,
        _ => en,
    }
}

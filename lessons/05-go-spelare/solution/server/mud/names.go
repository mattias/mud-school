package mud

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Errors that ValidateName can return. Defining errors as variables lets the
// caller compare with errors.Is(err, ErrNameTooShort).
var (
	ErrNameTooShort = errors.New(MsgErrNameTooShort)
	ErrNameTooLong  = errors.New(MsgErrNameTooLong)
	ErrNameInvalid  = errors.New(MsgErrNameInvalid)
	ErrNameTaken    = errors.New(MsgErrNameTaken)
)

// ValidateName checks that a name is 2–16 letters (å, ä, ö and Japanese
// characters count as letters; digits and spaces do not).
func ValidateName(name string) error {
	n := utf8.RuneCountInString(name)
	switch {
	case n < 2:
		return ErrNameTooShort
	case n > 16:
		return ErrNameTooLong
	}
	for _, r := range name {
		if !unicode.IsLetter(r) {
			return ErrNameInvalid
		}
	}
	return nil
}

// NormalizeName makes the first letter upper case and the rest lower case: "aNNA" -> "Anna".
func NormalizeName(name string) string {
	first, size := utf8.DecodeRuneInString(name)
	if first == utf8.RuneError {
		return name
	}
	return string(unicode.ToUpper(first)) + strings.ToLower(name[size:])
}

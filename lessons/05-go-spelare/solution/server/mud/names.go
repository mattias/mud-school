package mud

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Fel som ValidateName kan returnera. Att definiera fel som variabler gör
// att anroparen kan jämföra med errors.Is(err, ErrNameTooShort).
var (
	ErrNameTooShort = errors.New("namnet måste vara minst 2 tecken")
	ErrNameTooLong  = errors.New("namnet får vara högst 16 tecken")
	ErrNameInvalid  = errors.New("namnet får bara innehålla bokstäver")
	ErrNameTaken    = errors.New("det namnet är redan upptaget")
)

// ValidateName kontrollerar att ett namn är 2–16 bokstäver (å, ä och ö räknas
// som bokstäver, siffror och mellanslag gör det inte).
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

// NormalizeName gör första bokstaven stor och resten små: "aNNA" -> "Anna".
func NormalizeName(name string) string {
	first, size := utf8.DecodeRuneInString(name)
	if first == utf8.RuneError {
		return name
	}
	return string(unicode.ToUpper(first)) + strings.ToLower(name[size:])
}

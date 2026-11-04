package mud

import (
	"errors"
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
	// TODO(lesson 5)
	return nil
}

// NormalizeName makes the first letter upper case and the rest lower case: "aNNA" -> "Anna".
func NormalizeName(name string) string {
	// TODO(lesson 5)
	return name
}

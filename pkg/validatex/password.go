package validatex

import (
	"strings"
	"unicode"
)

func ValidatePassword(password string) bool {
	if !passwordLengthValid(password) || strings.Contains(password, " ") {
		return false
	}

	var (
		hasLetter bool
		hasDigit  bool
	)

	for _, char := range password {
		switch {
		case unicode.IsLetter(char):
			hasLetter = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	return hasLetter && hasDigit
}

func passwordLengthValid(password string) bool {
	return len(password) >= 6 && len(password) < 255
}

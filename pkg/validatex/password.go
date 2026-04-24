package validatex

import (
	"strings"
	"unicode"
)

func ValidatePassword(password string) bool {
	if len(password) < 6 || len(password) >= 255 {
		return false
	}

	if strings.Contains(password, " ") {
		return false
	}

	var hasLetter bool
	
	var hasDigit bool

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
package domain

import (
	"regexp"
	"strings"
	"time"
	"unicode"
)

type User struct {
	ID               int64
	Email            string
	Password         string
	Birthdate        *time.Time
	AvatarFileKey    *string
	RegistrationDate time.Time
	IsActive         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func Validate(email, password string) bool {
	return validateEmail(email) && validatePassword(password)
}

func ValidatePassword(password string) bool {
	return validatePassword(password)
}

func validateEmail(email string) bool {
	// Проверка длины (max 64)
	if len(email) == 0 || len(email) > 64 {
		return false
	}

	// Основное регулярное выражение для email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return false
	}

	// Проверка наличия @
	if !strings.Contains(email, "@") {
		return false
	}

	// Проверка что после @ есть хотя бы одна точка
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	domain := parts[1]
	// Проверяем что в домене нет двух точек подряд
	if strings.Contains(domain, "..") {
		return false
	}

	// Проверяем что есть хотя бы одна точка в домене
	if !strings.Contains(domain, ".") {
		return false
	}

	return true
}

func validatePassword(password string) bool {
	// 6 <= длина <= 255
	if len(password) < 6 || len(password) >= 255 {
		return false
	}

	// Проверка на пробелы
	if strings.Contains(password, " ") {
		return false
	}

	var (
		hasLetter bool
		hasDigit  bool
	)

	// Проверка каждого символа
	for _, char := range password {
		switch {
		case unicode.IsLetter(char):
			hasLetter = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	// должна быть хотя бы одна буква и одна цифра
	return hasLetter && hasDigit
}

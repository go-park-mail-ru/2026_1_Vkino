package domain

import (
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

type User struct {
	ID               uuid.UUID
	Email            string
	Password         string
	RegistrationDate time.Time
	IsActive         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}


func (u *User) Name() string {
	return "users"
}


func (u *User) Validate() bool {
	return u.validateEmail() && u.validatePassword()
}


func (u *User) validateEmail() bool {
	// Проверка длины (max 64)
	if len(u.Email) == 0 || len(u.Email) > 64 {
		return false
	}
	
	// Основное регулярное выражение для email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return false
	}
	
	// Проверка наличия @
	if !strings.Contains(u.Email, "@") {
		return false
	}
	
	// Проверка что после @ есть хотя бы одна точка
	parts := strings.Split(u.Email, "@")
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


func (u *User) validatePassword() bool {
	// 6 <= длина <= 255
	if len(u.Password) < 6 || len(u.Password) >= 255 {
		return false
	}
	
	// Проверка на пробелы
	if strings.Contains(u.Password, " ") {
		return false
	}
	
	var (
		hasLetter bool
		hasDigit  bool
	)
	
	// Проверка каждого символа
	for _, char := range u.Password {
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
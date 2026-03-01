package auth

import (
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/pkg/db"
)

func TestUserRepository(t *testing.T) {
	db := db.NewDB([]db.Named{&User{}})
	repo := NewRepository(db)

	login := "testuser"
	password := "password123"

	_, err := repo.CreateUser(login, password)

	if err != nil {
		t.Fatalf("Error creating user: %s", err)
	}

	user, err := repo.GetUserByLogin(login)

	if user.Password != password {
		t.Fatal("Passwords don't match")
	}

	err = repo.DeleteUser(login)

	if err != nil {
		t.Fatalf("Error deleting user: %s", err)
	}

	user, err = repo.GetUserByLogin(login)

	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("No user not found error while getting user after deleting")
	}
}

package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func setupTestRepo(t *testing.T) *Repository {
	models := []Named{&User{}}
	repo := InitRepo(models)
	return repo
}

func TestCreateAndGetUser(t *testing.T) {
	repo := setupTestRepo(t)

	login := "testuser"
	password := "password123"

	user, err := repo.CreateUser(login, password)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.Login != login {
		t.Errorf("Expected login %s, got %s", login, user.Login)
	}
	if user.Password != password {
		t.Errorf("Expected password %s, got %s", password, user.Password)
	}
	if user.ID == uuid.Nil {
		t.Error("Expected non-zero UUID")
	}
	if !user.IsActive {
		t.Error("Expected user to be active")
	}
	if user.RegistrationDate.IsZero() {
		t.Error("Expected registration date to be set")
	}
	if user.CreatedAt.IsZero() {
		t.Error("Expected created at to be set")
	}
	if user.UpdatedAt.IsZero() {
		t.Error("Expected updated at to be set")
	}

	retrievedUser, err := repo.GetUserByLogin(login)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if retrievedUser.ID != user.ID {
		t.Errorf("Expected ID %v, got %v", user.ID, retrievedUser.ID)
	}
}

func TestCreateDuplicateUser(t *testing.T) {
	repo := setupTestRepo(t)

	login := "duplicate"
	password := "pass123"

	_, err := repo.CreateUser(login, password)
	if err != nil {
		t.Fatalf("Failed to create first user: %v", err)
	}

	_, err = repo.CreateUser(login, "anotherpass")
	if err != ErrUserAlreadyExists {
		t.Errorf("Expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestGetNonExistentUser(t *testing.T) {
	repo := setupTestRepo(t)

	_, err := repo.GetUserByLogin("nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestUpdateUser(t *testing.T) {
	repo := setupTestRepo(t)

	login := "updateuser"
	oldPassword := "oldpass"
	newPassword := "newpass"

	user, err := repo.CreateUser(login, oldPassword)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	oldUpdatedAt := user.UpdatedAt
	oldRegistrationDate := user.RegistrationDate
	oldCreatedAt := user.CreatedAt

	time.Sleep(time.Millisecond)

	updatedUser, err := repo.UpdateUser(login, newPassword)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	if updatedUser.Password != newPassword {
		t.Errorf("Expected password %s, got %s", newPassword, updatedUser.Password)
	}
	if updatedUser.Login != login {
		t.Errorf("Expected login %s, got %s", login, updatedUser.Login)
	}
	if updatedUser.ID != user.ID {
		t.Errorf("Expected ID %v, got %v", user.ID, updatedUser.ID)
	}
	if !updatedUser.UpdatedAt.After(oldUpdatedAt) {
		t.Error("Expected UpdatedAt to be after old UpdatedAt")
	}
	if !updatedUser.RegistrationDate.Equal(oldRegistrationDate) {
		t.Error("RegistrationDate should not change")
	}
	if !updatedUser.CreatedAt.Equal(oldCreatedAt) {
		t.Error("CreatedAt should not change")
	}
	if !updatedUser.IsActive {
		t.Error("IsActive should remain true")
	}

	retrievedUser, err := repo.GetUserByLogin(login)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if retrievedUser.Password != newPassword {
		t.Errorf("Expected stored password %s, got %s", newPassword, retrievedUser.Password)
	}
}

func TestUpdateNonExistentUser(t *testing.T) {
	repo := setupTestRepo(t)

	_, err := repo.UpdateUser("nonexistent", "newpass")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	repo := setupTestRepo(t)

	login := "deleteuser"

	user, err := repo.CreateUser(login, "pass")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	oldUpdatedAt := user.UpdatedAt
	time.Sleep(time.Millisecond)

	err = repo.DeleteUser(login)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}

	deletedUser, err := repo.GetUserByLogin(login)
	if err != nil {
		t.Fatalf("Failed to get deleted user: %v", err)
	}

	if deletedUser.IsActive {
		t.Error("Expected user to be inactive after delete")
	}
	if !deletedUser.UpdatedAt.After(oldUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated after delete")
	}
}

func TestDeleteNonExistentUser(t *testing.T) {
	repo := setupTestRepo(t)

	err := repo.DeleteUser("nonexistent")
	if err != ErrUserNotFound {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
	}
}

func TestGetUsers(t *testing.T) {
	repo := setupTestRepo(t)

	users := []struct {
		login    string
		password string
	}{
		{"user1", "pass1"},
		{"user2", "pass2"},
		{"user3", "pass3"},
		{"user4", "pass4"},
	}

	for _, u := range users {
		_, err := repo.CreateUser(u.login, u.password)
		if err != nil {
			t.Fatalf("Failed to create user %s: %v", u.login, err)
		}
	}

	repo.DeleteUser("user3")

	activeUsers, err := repo.GetUsers()
	if err != nil {
		t.Fatalf("Failed to get users: %v", err)
	}

	if len(activeUsers) != 3 {
		t.Errorf("Expected 3 active users, got %d", len(activeUsers))
	}

	for _, u := range activeUsers {
		if u.Login == "user3" {
			t.Error("Deleted user should not be in active users list")
		}
	}
}

func TestUserSerialization(t *testing.T) {
	repo := setupTestRepo(t)

	original, err := repo.CreateUser("serialize_test", "pass")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	repo.mu.RLock()
	data := repo.tbls["users"]["serialize_test"]
	repo.mu.RUnlock()

	var deserialized User
	err = UserDeserialize(data, &deserialized)
	if err != nil {
		t.Fatalf("Failed to deserialize: %v", err)
	}

	if deserialized.ID != original.ID {
		t.Errorf("ID mismatch: expected %v, got %v", original.ID, deserialized.ID)
	}
	if deserialized.Login != original.Login {
		t.Errorf("Login mismatch: expected %s, got %s", original.Login, deserialized.Login)
	}
	if deserialized.Password != original.Password {
		t.Errorf("Password mismatch: expected %s, got %s", original.Password, deserialized.Password)
	}
	if !deserialized.RegistrationDate.Equal(original.RegistrationDate) {
		t.Error("RegistrationDate mismatch")
	}
}

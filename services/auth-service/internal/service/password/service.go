package password

import "golang.org/x/crypto/bcrypt"

type Service interface {
	Hash(password string) (string, error)
	Compare(hash string, rawPassword string) error
}

type BcryptService struct{}

func New() *BcryptService {
	return &BcryptService{}
}

func (s *BcryptService) Hash(password string) (string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(passwordHash), nil
}

func (s *BcryptService) Compare(hash string, rawPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(rawPassword))
}
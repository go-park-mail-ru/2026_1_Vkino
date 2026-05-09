package usecase

type service struct{}

func New() Usecase {
	return &service{}
}

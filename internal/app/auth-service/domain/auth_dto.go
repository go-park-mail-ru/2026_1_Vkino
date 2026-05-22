package domain

import jwtsvc "github.com/go-park-mail-ru/2026_1_VKino/pkg/service/jwt"

type TokenPair struct {
	accessToken  string
	refreshToken string
}

type AuthContext = jwtsvc.AuthContext

func NewTokenPair(accessToken, refreshToken string) TokenPair {
	return TokenPair{
		accessToken:  accessToken,
		refreshToken: refreshToken,
	}
}

func (p TokenPair) AccessToken() string {
	return p.accessToken
}

func (p TokenPair) RefreshToken() string {
	return p.refreshToken
}

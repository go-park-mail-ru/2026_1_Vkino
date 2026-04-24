package domain

import jwtsvc "github.com/go-park-mail-ru/2026_1_VKino/pkg/service/jwt"

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthContext = jwtsvc.AuthContext

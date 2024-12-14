package user

import (
	"context"
	def "go-photo/pkg/account_v1"
)

func (s *service) Login(ctx context.Context, login string, password string) (string, error) {
	// TODO encytped password
	resp, err := s.accountClient.Login(ctx, &def.LoginRequest{Email: login, EncryptedPassword: password})
	if err != nil {
		return "", err
	}

	return resp.JwtToken, nil
}

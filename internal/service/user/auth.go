package user

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"go-photo/internal/config"
	serviceErr "go-photo/internal/service/error"
	serviceUserModel "go-photo/internal/service/user/model"
	def "go-photo/pkg/account_v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

func (s *service) Login(ctx context.Context, email string, password string) (string, error) {
	publicKey, err := s.getPublicKey(ctx)
	if err != nil {
		return "", s.handleGRPCErr(err)
	}

	encryptedPassword, err := s.utils.EncryptPassword(publicKey, password)
	if err != nil {
		return "", fmt.Errorf("%w: %v", serviceErr.InternalError, err)
	}

	resp, err := s.accountClient.Login(ctx, &def.LoginRequest{Email: email, EncryptedPassword: encryptedPassword})
	if err != nil {
		return "", s.handleGRPCErr(err)
	}

	return resp.JwtToken, nil
}

func (s *service) Register(ctx context.Context, input serviceUserModel.RegisterParams) (serviceUserModel.RegisterInfo, error) {
	publickKey, err := s.getPublicKey(ctx)
	if err != nil {
		return serviceUserModel.RegisterInfo{}, s.handleGRPCErr(err)
	}

	encryptedPassword, err := s.utils.EncryptPassword(publickKey, input.Password)
	if err != nil {
		return serviceUserModel.RegisterInfo{}, fmt.Errorf("%w: %v", serviceErr.InternalError, err)
	}

	resp, err := s.accountClient.Signup(ctx, &def.CreateRequest{
		Email:             input.Email,
		EncryptedPassword: encryptedPassword,
	})
	if err != nil {
		return serviceUserModel.RegisterInfo{}, s.handleGRPCErr(err)
	}

	info := serviceUserModel.RegisterInfo{
		UserUUID: resp.Uuid,
		Token:    resp.JwtToken,
	}

	return info, nil
}

func (s *service) handleGRPCErr(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return serviceErr.InternalError
	}

	switch st.Code() {
	case codes.NotFound:
		return serviceErr.UserNotFoundError
	case codes.AlreadyExists:
		return serviceErr.UserAlreadyExistsError
	}

	return fmt.Errorf("%w: %v", serviceErr.InternalError, st.Message())
}

func (s *service) getPublicKey(ctx context.Context) (*string, error) {
	s.publicKeyCache.mu.RLock()
	if time.Now().Before(s.publicKeyCache.ttl) && s.publicKeyCache.key != "" {
		s.publicKeyCache.mu.RUnlock()
		log.Infof("public key from cache: %s", s.publicKeyCache.key)
		return &s.publicKeyCache.key, nil
	}
	s.publicKeyCache.mu.RUnlock()

	publicKey, err := s.accountClient.GetPublicKey(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, s.handleGRPCErr(err)
	}

	s.publicKeyCache.mu.Lock()
	s.publicKeyCache.key = publicKey.PublicKey
	s.publicKeyCache.ttl = time.Now().Add(config.RSAPublicKeyDefaultTTL)
	s.publicKeyCache.mu.Unlock()

	return &publicKey.PublicKey, nil
}

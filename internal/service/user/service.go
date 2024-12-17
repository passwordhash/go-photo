package user

import (
	def "go-photo/internal/service"
	"go-photo/internal/utils"
	desc "go-photo/pkg/account_v1"
	"sync"
	"time"
)

// Проверка на соответствие интерфейсу UserService (для статической проверки)
var _ def.UserService = (*service)(nil)

type service struct {
	accountClient desc.AccountServiceClient

	publicKeyCache publicKeyCache

	utils utils.Inteface
}

func NewService(accountClient desc.AccountServiceClient, u utils.Inteface) *service {
	if u == nil {
		u = utils.New()
	}
	return &service{
		accountClient: accountClient,
		utils:         u,
	}
}

type publicKeyCache struct {
	mu  sync.RWMutex
	key string
	ttl time.Time
}

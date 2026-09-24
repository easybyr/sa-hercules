package service

import (
	"github.com/example/sa-hercules/admin/internal/repository"
	"github.com/example/sa-hercules/admin/pkg/token"
)

// Service 包含后台管理系统业务逻辑。
type Service struct {
	store  *repository.Store
	tokens *token.Manager
}

// New 创建业务服务。
func New(store *repository.Store, tokenManager *token.Manager) *Service {
	return &Service{store: store, tokens: tokenManager}
}

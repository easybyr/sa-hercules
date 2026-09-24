package service

import "context"

// Health 检查数据库连接。
func (s *Service) Health(ctx context.Context) error {
	_, err := s.store.DB().GetValue(ctx, "SELECT 1")
	return err
}

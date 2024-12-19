package photo

import (
	"context"
	"fmt"
)

func (s *service) DeletePhoto(ctx context.Context, photoID int) error {
	_, err := s.photoRepository.DeletePhoto(ctx, photoID)
	if err != nil {
		return fmt.Errorf("failed to delete photo: %w", err)
	}
	return nil
}

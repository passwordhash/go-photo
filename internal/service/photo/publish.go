package photo

import (
	"context"
)

func (s *service) PublishPhoto(ctx context.Context, userUUID string, photoID int) (string, error) {
	photo, err := s.getUserPhoto(ctx, userUUID, photoID)
	if err != nil {
		return "", err
	}

	publicToken, err := s.photoRepository.CreatePhotoPublishedInfo(ctx, photo.ID)
	if err := s.HandleRepoErr(err); err != nil {
		return "", err
	}

	return publicToken, nil
}

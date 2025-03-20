package photo

import (
	"context"
)

func (s *service) UnpublishPhoto(ctx context.Context, userUUID string, photoID int) error {
	photo, err := s.getUserPhoto(ctx, userUUID, photoID)
	if err != nil {
		return err
	}

	err = s.photoRepository.DeletePhotoPublishedInfo(ctx, photo.ID)
	handledErr := s.HandleError(err)

	return handledErr
}

package service

import (
	"context"
	"go-photo/internal/model"
	serviceAuthModel "go-photo/internal/service/auth/model"
	servicePhotoModel "go-photo/internal/service/photo/model"
	"mime/multipart"
)

//go:generate mockgen -destination=mock/mocks.go -source=interface.go

// type TokenService interface {
// VerifyToken проверяет токен и возвращает payload из токена
// VerifyToken(ctx context.Context, token string) (serviceUserModel.TokenPayload, error)
// }

type AuthService interface {
	// TODO: doc
	Register(
		ctx context.Context,
		params serviceAuthModel.RegisterParams,
	) (userUUID string, err error)

	// Login выполняет аутентификацию пользователя по логину и паролю. Возвращает JWT token
	Login(ctx context.Context, email string, password string) (jwtToken string, err error)
}

//type UserService interface {
// // Login выполняет аутентификацию пользователя по логину и паролю. Возвращает JWT token
// Login(ctx context.Context, login string, password string) (string, error)
// // Register регистрирует нового пользователя.
// Register(ctx context.Context, input serviceUserModel.RegisterParams) (serviceUserModel.RegisterInfo, error)
// Get(ctx context.Context, uuid string) (model.User, error)
//}

type PhotoService interface {
	// UploadPhoto загружает фотографию и сохраняет ее в файловой системе и базе данных.
	// Возвращает ID загруженной фотографии.
	UploadPhoto(ctx context.Context, userUUID string, photoFile *multipart.FileHeader) (int, error)

	// UploadBatchPhotos загружает несколько фотографий конкурентно. Возвращает список информации о загруженных фотографиях.
	// Если возникла ошибка во время загрузки фотографии, то прикрепляет информацию об ошибке.
	UploadBatchPhotos(ctx context.Context, userUUID string, photoFiles []*multipart.FileHeader) (*servicePhotoModel.UploadInfoList, error)

	// PublishPhoto публикует фотографию, делая ее доступной для других пользователей.
	// Осуществляет проверку прав доступа к фотографии.
	PublishPhoto(ctx context.Context, userUUID string, photoID int) (string, error)

	// GetPhotoVersions получает все версии фотографии по ее ID.
	// Осуществляет проверку прав доступа к фотографии.
	// Возвращает список версий фотографии.
	GetPhotoVersions(ctx context.Context, userUUID string, photoID int) ([]model.PhotoVersion, error)

	// GetPhotoFileByVersionAndToken получает файл публичной фотографии по ее версии и токену.
	// TODO: tests
	GetPhotoFileByVersionAndToken(ctx context.Context, token string, version string) ([]byte, error)

	// UnpublishPhoto отменяет публикацию фотографии, делая ее недоступной для других пользователей.
	// Осуществляет проверку прав доступа к фотографии.
	UnpublishPhoto(ctx context.Context, userUUID string, photoID int) error

	// HandleRepoErr обрабатывает ошибки, возвращаемые репозиторием.
	// Обрабатывает ошибки:
	// - NotFoundError
	// - ConflictError
	// Если ошибка не распознана, возвращает UnexpectedError.
	HandleRepoErr(err error) error
}

package converter

import (
	"go-photo/internal/service/user/model"
	def "go-photo/pkg/account_v1"
)

func ToTokenPayloadFromProto(repsone *def.VerifyTokenResponse) *model.TokenPayload {
	return &model.TokenPayload{
		UserUUID: repsone.Uuid,
	}
}

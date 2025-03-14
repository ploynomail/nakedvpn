package biz

import (
	"NakedVPN/internal/conf"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type HandleClientUseCase struct {
	conf *conf.Client
	log  *log.Helper
}

func NewHandleClientUseCase(conf *conf.Client, logger log.Logger) *HandleClientUseCase {
	return &HandleClientUseCase{
		conf: conf,
		log:  log.NewHelper(log.With(logger, "module", "biz/handle")),
	}
}

func (h *HandleClientUseCase) HandleCommandReqAuth(data []byte) ([]byte, *errors.Error) {
	return []byte(h.conf.Config.AuthKey), nil
}

func (h *HandleClientUseCase) HandleCommandAuthResult(data []byte) (bool, *errors.Error) {
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return false, ErrInvalidData
	}
	if resp.Code != 0 {
		return false, ErrAuthFailed
	}
	return true, nil
}

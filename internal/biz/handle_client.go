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

func (h *HandleClientUseCase) HandleCommandReqAuth(data []byte) ([]byte, error) {
	simpleCodec := SimpleCodec{
		CurrentOrganize: uint16(h.conf.Config.Organize),
		CommandCode:     CommandAuth,
		Data:            []byte(h.conf.Config.AuthKey),
	}
	res, err := simpleCodec.Encode()
	if err != nil {
		h.log.Errorf("simpleCodec.Encode: %v", err)
		return nil, err
	}
	return res, nil
}

func (h *HandleClientUseCase) HandleCommandAuthResult(data []byte) (bool, *errors.Error) {
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return false, ErrInvalidData
	}
	if resp.Code != 911 {
		return false, ErrAuthFailed
	}
	return true, nil
}

package service

import (
	"NakedVPN/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/panjf2000/gnet/v2"
)

type StreamProcessing struct {
	orgUc *biz.OrganizeUseCase
	log   *log.Helper
}

func NewStreamProcessing(orgUc *biz.OrganizeUseCase, logger log.Logger) *StreamProcessing {
	return &StreamProcessing{
		orgUc: orgUc,
		log:   log.NewHelper(log.With(logger, "module", "biz/stream_processing")),
	}
}

func (s *StreamProcessing) ReqAuth(c gnet.Conn) ([]byte, error) {
	var simpleCodec SimpleCodec = SimpleCodec{}
	simpleCodec.CurrentOrganize = 0
	simpleCodec.CommandCode = uint16(biz.CommandReqAuth)
	simpleCodec.Data = []byte("req auth")
	data, err := simpleCodec.Encode()
	if err != nil {
		return nil, biz.ErrIncompletePacket
	}
	return data, nil
}

func (s *StreamProcessing) Processing(c gnet.Conn) error {
	var simpleCodec SimpleCodec = SimpleCodec{}
	if err := simpleCodec.Decode(c); err != nil {
		return biz.ErrIncompletePacket
	}
	switch simpleCodec.CommandCode {
	case uint16(biz.CommandReqAuth):
		// 处理认证请求
		s.log.Infof("Processing CommandReqAuth")
	case uint16(biz.CommandData):
		// 处理数据
	}
	return nil
}

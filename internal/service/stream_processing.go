package service

import (
	"NakedVPN/internal/biz"
	"encoding/json"
	"io"
	"net"
	"syscall"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/panjf2000/gnet/v2"
)

type StreamProcessing struct {
	orgUc    *biz.OrganizeUseCase
	handleUc *biz.HandleUseCase
	log      *log.Helper
}

func NewStreamProcessing(orgUc *biz.OrganizeUseCase, handleUc *biz.HandleUseCase, logger log.Logger) *StreamProcessing {
	return &StreamProcessing{
		orgUc:    orgUc,
		handleUc: handleUc,
		log:      log.NewHelper(log.With(logger, "module", "biz/stream_processing")),
	}
}

func (s *StreamProcessing) ReqAuth(c gnet.Conn) []byte {
	data := s.handleUc.HandleCommandReqAuth()
	return data
}

func (s *StreamProcessing) Close(c gnet.Conn) {
	// TODO:释放客户端ip资源
}

func (s *StreamProcessing) Processing(c gnet.Conn) error {
	var simpleCodec biz.SimpleCodec = biz.SimpleCodec{}
	if err := simpleCodec.Decode(c); err != nil {
		if err == io.EOF || err == net.ErrClosed || err == syscall.EPIPE {
			return nil
		}
		s.log.Errorf("simpleCodec.DecodeForStdNet: %v", err)
	}
	switch simpleCodec.CommandCode {
	case biz.CommandAuth:
		// 为该组织准备虚拟tun设备
		if err := s.orgUc.PrepareTun(simpleCodec.CurrentOrganize); err != nil {
			s.RespError(c, err, simpleCodec.CurrentOrganize)
			return nil
		}
		// 处理认证请求
		s.log.Debugf("Processing CommandAuth")
		data, err := s.handleUc.HandleCommandAuth(simpleCodec.CurrentOrganize, simpleCodec.Data, c)
		if err != nil {
			s.RespError(c, err, simpleCodec.CurrentOrganize)
			return nil
		}

		// 返回认证成功
		return s.Resp(c, data)

	case biz.CommandData:
		// 处理数据
		s.log.Infof("Processing CommandData")
		if err := s.handleUc.HandleCommandData(simpleCodec.CurrentOrganize, simpleCodec.Data); err != nil {
			s.RespError(c, err, simpleCodec.CurrentOrganize)
		}
	case biz.CommandError:
		// 处理错误
		s.log.Infof("Processing CommandError")
	}
	return nil
}

func (s *StreamProcessing) Resp(c gnet.Conn, data []byte) error {
	_, err := c.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (s *StreamProcessing) RespError(c gnet.Conn, err error, org uint16) {
	s.log.Errorf("RespError: %v", err)
	var resp biz.Response = biz.Response{
		Code: errors.Code(err),
		Data: nil,
		Msg:  errors.Reason(err),
		Ts:   time.Now().Unix(),
	}
	respJson, e := json.Marshal(resp)
	if e != nil {
		s.log.Errorf("RespError json.Marshal: %v", err)
		return
	}
	var simpleCodec biz.SimpleCodec = biz.SimpleCodec{
		CurrentOrganize: uint16(org),
		CommandCode:     biz.CommandError,
		Data:            respJson,
	}
	data, e := simpleCodec.Encode()
	if e != nil {
		s.log.Errorf("RespError simpleCodec.Encode: %v", err)
		return
	}
	_, e = c.Write(data)
	if e != nil {
		s.log.Errorf("RespError c.Write: %v", err)
		return
	}
}

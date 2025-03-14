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
		if err == io.EOF || err == net.ErrClosed || err == syscall.EPIPE {
			return nil
		}
		s.log.Errorf("simpleCodec.DecodeForStdNet: %v", err)
	}
	switch simpleCodec.CommandCode {
	case uint16(biz.CommandAuth):
		// 处理认证请求
		s.log.Infof("Processing CommandAuth")
		ok, err := s.handleUc.HandleCommandAuth(simpleCodec.CurrentOrganize, simpleCodec.Data)
		if err != nil {
			s.RespError(c, err, simpleCodec.CurrentOrganize)
		}
		if ok {
			simpleCodec.CommandCode = uint16(biz.CommandAuthResult)
			var resp biz.Response = biz.Response{
				Code: 0,
				Data: nil,
				Msg:  "success",
				Ts:   time.Now().Unix(),
			}
			respJson, e := json.Marshal(resp)
			if e != nil {
				s.RespError(c, e, simpleCodec.CurrentOrganize)
			}
			simpleCodec.Data = respJson
			data, err := simpleCodec.Encode()
			if err != nil {
				s.RespError(c, err, simpleCodec.CurrentOrganize)
			}
			err = s.Resp(c, data)
			if err != nil {
				s.RespError(c, err, simpleCodec.CurrentOrganize)
			}
		} else {
			simpleCodec.CommandCode = uint16(biz.CommandAuthResult)
			var resp biz.Response = biz.Response{
				Code: 1,
				Data: nil,
				Msg:  "auth failed",
				Ts:   time.Now().Unix(),
			}
			respJson, e := json.Marshal(resp)
			if e != nil {
				s.RespError(c, e, simpleCodec.CurrentOrganize)
			}
			simpleCodec.Data = respJson
			data, err := simpleCodec.Encode()
			if err != nil {
				s.RespError(c, err, simpleCodec.CurrentOrganize)
			}
			err = s.Resp(c, data)
			if err != nil {
				s.RespError(c, err, simpleCodec.CurrentOrganize)
			}
		}
	case uint16(biz.CommandData):
		// 处理数据
		s.log.Infof("Processing CommandData")
		if err := s.handleUc.HandleCommandData(simpleCodec.Data); err != nil {
			s.RespError(c, err, simpleCodec.CurrentOrganize)
		}
		// test echo
		var sc SimpleCodec = SimpleCodec{}
		sc.CurrentOrganize = simpleCodec.CurrentOrganize
		sc.CommandCode = uint16(biz.CommandData)
		sc.Data = simpleCodec.Data

		data, _ := sc.Encode()
		err := s.Resp(c, data)
		if err != nil {
			s.RespError(c, err, sc.CurrentOrganize)
		}
	case uint16(biz.CommandError):
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
	}
	var simpleCodec SimpleCodec = SimpleCodec{
		CurrentOrganize: uint16(org),
		CommandCode:     uint16(biz.CommandError),
		Data:            respJson,
	}
	data, e := simpleCodec.Encode()
	if e != nil {
		s.log.Errorf("RespError simpleCodec.Encode: %v", err)
	}
	_, e = c.Write(data)
	if e != nil {
		s.log.Errorf("RespError c.Write: %v", err)
	}
}

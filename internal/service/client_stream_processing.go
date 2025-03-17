package service

import (
	"NakedVPN/internal/biz"
	"NakedVPN/internal/conf"
	"encoding/json"
	"io"
	"net"
	"syscall"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
)

type ClientStreamProcessing struct {
	conf     *conf.Client
	handleUc *biz.HandleClientUseCase
	log      *log.Helper
}

func NewClientStreamProcessing(handleUc *biz.HandleClientUseCase, conf *conf.Client, logger log.Logger) *ClientStreamProcessing {
	return &ClientStreamProcessing{
		handleUc: handleUc,
		conf:     conf,
		log:      log.NewHelper(log.With(logger, "module", "biz/stream_processing")),
	}
}

func (s *ClientStreamProcessing) Processing(c net.Conn) error {
	for {
		commandCode, body, err := biz.Unpack(c)
		if err != nil {
			if err == io.EOF || err == net.ErrClosed || err == syscall.EPIPE {
				return nil
			}
			if oe, ok := err.(*net.OpError); ok && oe.Op == "read" {
				return nil
			}
			s.log.Errorf("Processing Unpack: %v", err)
			continue
		}
		switch commandCode {
		case biz.CommandReqAuth:
			// 处理认证请求
			s.log.Infof("Processing CommandReqAuth")
			data, err := s.handleUc.HandleCommandReqAuth(body)
			if err != nil {
				s.log.Errorf("Processing CommandReqAuth: %v", err)
				s.RespError(c, err)
				continue
			}
			s.Resp(c, data)
		case biz.CommandAuthResult:
			// 处理认证请求
			s.log.Infof("Processing CommandAuthResult")
			ok, err := s.handleUc.HandleCommandAuthResult(body, c)
			if err != nil || !ok {
				if errors.Is(err, biz.ErrCreateTunFailed) {
					return err
				}
				s.log.Errorf("Processing CommandAuthResult: %v", err)
				return err
			}
		case biz.CommandData:
			// 处理数据
			s.log.Infof("Processing CommandData %s", string(body))
		case biz.CommandError:
			// 处理错误
			s.log.Infof("Processing CommandError")
		}
	}
}

func (s *ClientStreamProcessing) Resp(c net.Conn, data []byte) error {
	_, err := c.Write(data)
	if err != nil {
		s.log.Errorf("Resp: %v", err)
		return err
	}
	return nil
}

func (s *ClientStreamProcessing) RespError(c net.Conn, err error) {
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
	var simpleCodec biz.SimpleCodec = biz.SimpleCodec{
		CurrentOrganize: uint16(s.conf.Config.Organize),
		CommandCode:     biz.CommandError,
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

// func (s *ClientStreamProcessing) SendTest(c net.Conn) {
// 	i := 0
// 	for {
// 		i++
// 		var simpleCodec biz.SimpleCodec = biz.SimpleCodec{
// 			CurrentOrganize: uint16(s.conf.Config.Organize),
// 			CommandCode:     biz.CommandData,
// 			Data:            []byte("Echo"),
// 		}
// 		data, err := simpleCodec.Encode()
// 		if err != nil {
// 			s.RespError(c, err)
// 			continue
// 		}
// 		s.Resp(c, data)
// 		time.Sleep(1 * time.Second)
// 	}
// }

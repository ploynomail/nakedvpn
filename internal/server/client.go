package server

import (
	"NakedVPN/internal/conf"
	"NakedVPN/internal/service"
	"context"
	"net"

	"github.com/go-kratos/kratos/v2/log"
)

type NetClient struct {
	conf *conf.Client

	clientSp *service.ClientStreamProcessing
	c        net.Conn
	log      *log.Helper
}

func NewNetClient(conf *conf.Client, clientSp *service.ClientStreamProcessing, logger log.Logger) *NetClient {
	return &NetClient{
		conf:     conf,
		clientSp: clientSp,
		log:      log.NewHelper(log.With(logger, "module", "client/net")),
	}
}

func (s *NetClient) Start(context.Context) error {
	s.log.Infof("NetClient start")
	if err := s.Dial(context.Background()); err != nil {
		s.log.Errorf("NetClient dial error: %v", err)
		return err
	}
	if err := s.clientSp.Processing(s.c); err != nil {
		s.log.Errorf("NetClient processing error: %v", err)
		s.c.Close()
		return err
	}
	return nil
}

func (s *NetClient) Stop(context.Context) error {
	s.log.Infof("NetClient stop")
	s.c.Close()
	return nil
}

func (s *NetClient) Dial(context.Context) error {
	s.log.Infof("NetClient dial")

	c, err := net.Dial(s.conf.Target.Network, s.conf.Target.Addr)
	if err != nil {
		s.log.Errorf("dial error: %v", err)
		return err
	}
	s.c = c
	return nil
}

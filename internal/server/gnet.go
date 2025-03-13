package server

import (
	"NakedVPN/internal/conf"
	"NakedVPN/internal/service"
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/panjf2000/gnet/v2"
	"github.com/panjf2000/gnet/v2/pkg/pool/goroutine"
)

type NetServer struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	streamProcessingService *service.StreamProcessing
	conf                    *conf.Server

	Eng  gnet.Engine
	pool *goroutine.Pool
	log  *log.Helper
	gnet.BuiltinEventEngine
}

func NewNetServer(sp *service.StreamProcessing, conf *conf.Server, logger log.Logger) *NetServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &NetServer{
		ctx:                     ctx,
		ctxCancel:               cancel,
		streamProcessingService: sp,
		conf:                    conf,
		pool:                    goroutine.Default(),
		log:                     log.NewHelper(log.With(logger, "module", "server/net")),
	}
}

func (s *NetServer) Start(context.Context) error {
	var err error
	var errChan chan error = make(chan error, 1)
	go func() {
		err = gnet.Run(s,
			s.conf.Gnet.Network+"://"+s.conf.Gnet.Addr,
			gnet.WithMulticore(s.conf.Gnet.Multicore),
			gnet.WithLogger(s.log),
			gnet.WithReusePort(true),
			gnet.WithTCPKeepAlive(time.Minute*5),
			gnet.WithTicker(true),
		)
		if err != nil {
			errChan <- err
		}
	}()
	select {
	case err = <-errChan:
		s.log.Errorf("[GNET] server start error: %v", err)
	case <-s.ctx.Done():
		s.log.Info("[GNET] server done")
	}
	return nil
}

func (s NetServer) Stop(context.Context) error {
	s.log.Info("[GNET] server stopping")
	s.Eng.Stop(s.ctx)
	s.ctxCancel()
	s.pool.Release()
	s.log.Info("[GNET] server stopped")
	return nil
}

// 当引擎准备好接受连接时，OnBoot 触发。参数引擎包含信息和各种实用程序。
func (s NetServer) OnBoot(eng gnet.Engine) (action gnet.Action) {
	return
}

// OnShutdown 在引擎关闭时触发，它会在关闭后立即被调用所有事件循环和连接都已关闭。
// func (s NetServer) OnShutdown(eng gnet.Engine) {
// 	s.log.Info("[GNet] server stopped")
// }

// 当打开新连接时，OnOpen 会触发。Conn c 包含有关连接的信息，例如其本地和远程地址。
// 参数 out 是要发送回远程的返回值。 通常不建议在 OnOpen 中将大量数据发送回远程。
func (s *NetServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	s.log.Debugf("[GNET] OnOpen: %v", c.RemoteAddr())
	data, err := s.streamProcessingService.ReqAuth(c)
	if err != nil {
		s.log.Errorf("[GNET] OnOpenAuth error: %v", err)
		return nil, gnet.Close
	}
	return data, gnet.None
}

// 当连接关闭时，OnClose 会触发。
// 参数 err 是最后一个已知的连接错误。
func (s NetServer) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	return gnet.Close
}

// 当套接字从远程接收数据时，OnTraffic 会触发。
// 注意:不允许将 Conn.Peek(int)/Conn.Next(int) 返回的 []byte 传递给新的 goroutine，
// 因为在 OnTraffic() 返回后，此 []byte 将在事件循环中重复使用。
// 如果您必须在新的 goroutine 中使用此 []byte，则应复制它或调用 Conn.Read([]byte)
// 将数据读入您自己的 []byte，然后将新的 []byte 传递给新的 goroutine。
func (s NetServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	// 消息路由
	if err := s.streamProcessingService.Processing(c); err != nil {
		s.log.Errorf("[GNET] OnTraffic error: %v", err)
		// TODO: 处理错误,不应该直接关闭连接
		return gnet.Close
	}
	return gnet.None
}

// OnTick 在引擎启动后立即触发，并将在延迟返回值指定的持续时间后再次触发,must set option gnet.WithTicker(true)
func (s NetServer) OnTick() (delay time.Duration, action gnet.Action) {
	return time.Second, gnet.None
}

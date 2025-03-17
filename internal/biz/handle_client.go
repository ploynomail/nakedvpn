package biz

import (
	"NakedVPN/internal/conf"
	"encoding/json"
	"net"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/songgao/water"
	"github.com/valyala/bytebufferpool"
)

type HandleClientUseCase struct {
	conf       *conf.Client
	receiveBuf *bytebufferpool.ByteBuffer
	sendBuf    *bytebufferpool.ByteBuffer
	log        *log.Helper
}

func NewHandleClientUseCase(conf *conf.Client, logger log.Logger) *HandleClientUseCase {
	return &HandleClientUseCase{
		conf:       conf,
		receiveBuf: bytebufferpool.Get(),
		sendBuf:    bytebufferpool.Get(),
		log:        log.NewHelper(log.With(logger, "module", "biz/handle")),
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

func (h *HandleClientUseCase) HandleCommandAuthResult(data []byte, c net.Conn) (bool, *errors.Error) {
	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return false, ErrInvalidData
	}
	if resp.Code != 911 {
		return false, ErrAuthFailed
	}
	iface, err := water.New(water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name: "NakedVPN",
		},
	})
	if err != nil {
		h.log.Errorf("water.New: %v", err)
		return false, ErrCreateTunFailed
	}
	// 根据返回的路由表，配置网卡
	go func() {
		for {
			n, err := h.receiveBuf.ReadFrom(c)
			h.log.Debugf("ReadFrom: %d", n)
			if err != nil {
				h.log.Errorf("h.receiveBuf.ReadFrom: %v", err)
				continue
			}
			var data []byte
			var simpleCodec SimpleCodec = SimpleCodec{
				CurrentOrganize: uint16(h.conf.Config.Organize),
				CommandCode:     CommandData,
				Data:            data,
			}
			data, err = simpleCodec.Encode()
			if err != nil {
				h.log.Errorf("simpleCodec.Encode: %v", err)
				continue
			}
			ns, err := iface.Read(data)
			h.log.Debugf("ReadFrom: %d", ns)
			if err != nil {
				h.log.Errorf("iface.Read: %v", err)
				continue
			}
		}
	}()
	go func() {
		for {
			n, err := c.Read(h.sendBuf.B)
			h.log.Debugf("ReadFrom: %d", n)
			if err != nil {
				h.log.Errorf("c.Read: %v", err)
				continue
			}
			_, data, err := Unpack(c)
			if err != nil {
				h.log.Errorf("Unpack: %v", err)
				continue
			}
			h.receiveBuf.B = data
			ns, err := h.receiveBuf.WriteTo(c)
			h.log.Debugf("WriteTo: %d", ns)
			if err != nil {
				h.log.Errorf("h.receiveBuf.WriteTo: %v", err)
				continue
			}
		}
	}()
	return true, nil
}

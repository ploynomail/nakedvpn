package biz

import (
	"NakedVPN/internal/conf"
	"encoding/json"
	"net"
	"os/exec"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/songgao/water"
)

type HandleClientUseCase struct {
	conf  *conf.Client
	iface *water.Interface
	log   *log.Helper
}

func NewHandleClientUseCase(conf *conf.Client, logger log.Logger) *HandleClientUseCase {
	return &HandleClientUseCase{
		conf: conf,
		log:  log.NewHelper(log.With(logger, "module", "biz/handle-client")),
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
			Name:       "NakedVPN",
			MultiQueue: true,
		},
	})
	if err != nil {
		h.log.Errorf("water.New: %v", err)
		return false, ErrCreateTunFailed
	}
	ec := resp.Data.(map[string]interface{})
	// 设置IP地址
	cmd := exec.Command("ip", "addr", "add", ec["client_ip"].(string), "dev", iface.Name())
	if err := cmd.Run(); err != nil {
		h.log.Errorf("Failed to set IP address for interface: %v", err)
	}

	// 设置接口为UP状态
	cmd = exec.Command("ip", "link", "set", iface.Name(), "up")
	if err := cmd.Run(); err != nil {
		h.log.Errorf("Failed to set interface UP: %v", err)
	}

	// 设置路由
	cmd = exec.Command("ip", "route", "add", ec["virtual_gateway"].(string), "dev", iface.Name())
	if err := cmd.Run(); err != nil {
		h.log.Errorf("Failed to set route: %v", err)
	}

	h.iface = iface
	// 根据返回的路由表，配置网卡
	// 从网卡读取数据，发送到服务器
	go func() {
		for {
			buf := make([]byte, 65535)
			n, err := iface.Read(buf)
			if err != nil {
				h.log.Errorf("iface.Read: %v", err)
				continue
			}
			h.log.Debugf("read %d bytes from iface", n)
			simpleCodec := SimpleCodec{
				CurrentOrganize: uint16(h.conf.Config.Organize),
				CommandCode:     CommandData,
				Data:            buf[:n],
			}
			res, err := simpleCodec.Encode()
			if err != nil {
				h.log.Errorf("simpleCodec.Encode: %v", err)
				continue
			}
			h.log.Debugf("write %d bytes to server", n)
			_, err = c.Write(res)
			if err != nil {
				h.log.Errorf("c.Write: %v", err)
				continue
			}
		}
	}()
	return true, nil
}

func (h *HandleClientUseCase) HandleCommandData(data []byte) error {
	n, err := h.iface.Write(data)
	if err != nil {
		h.log.Errorf("iface.Write: %v", err)
	}
	h.log.Infof("write %d bytes to iface", n)
	return nil
}

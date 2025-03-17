package biz

import (
	"encoding/json"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/panjf2000/gnet/v2"
)

type Command uint16

const (
	CommandHeartbeat      Command = iota + 1 // Heartbeat
	CommandReqAuth                           // Request Auth
	CommandInfoCollect                       // Info Collect
	CommandInfoReport                        // Info Report
	CommandAuth                              // Auth
	CommandAuthResult                        // Auth Result
	CommandData                              // Data
	CommandRouteUpdate                       // Route Update
	CommandClose                             // Close
	CommandUpdateSoftware                    // Update Software

	CommandError
)

type HandleUseCase struct {
	log       *log.Helper
	clientUc  *ClientUseCase
	ifacePort *IfacePort
	orgUc     *OrganizeUseCase
}

func NewHandleUseCase(orgUc *OrganizeUseCase, clientUc *ClientUseCase, logger log.Logger, ifacePort *IfacePort) *HandleUseCase {
	return &HandleUseCase{
		clientUc:  clientUc,
		ifacePort: ifacePort,
		log:       log.NewHelper(log.With(logger, "module", "biz/handle")),
		orgUc:     orgUc,
	}
}

// HandleCommandReqAuth 发出认证请求
func (h *HandleUseCase) HandleCommandReqAuth() []byte {
	var simpleCodec SimpleCodec = SimpleCodec{
		CurrentOrganize: 0,
		CommandCode:     CommandReqAuth,
		Data:            []byte(""),
	}
	data, err := simpleCodec.Encode()
	if err != nil {
		h.log.Errorf("simpleCodec.Encode: %v", err)
	}
	return data
}

// HandleCommandAuth 处理认证请求
func (h *HandleUseCase) HandleCommandAuth(orgId uint16, data []byte, c gnet.Conn) ([]byte, error) {
	if orgId == 0 {
		return nil, ErrInvalidOrganize
	}
	if !h.orgUc.AuthAccessKey(orgId, string(data)) {
		return nil, ErrAuthFailed
	}
	clientVirtualIP, err := h.orgUc.AllocateIpForOrg(orgId)
	if err != nil {
		return nil, err
	}
	gw, err := h.orgUc.GetOrgVirtualGateway(orgId)
	if err != nil {
		return nil, err
	}
	var enforceClient EnforcementClient = EnforcementClient{
		ClientIP:       clientVirtualIP.String(),
		VirtualGateway: gw,
	}
	var resp Response = Response{
		Code: 911,
		Data: enforceClient,
		Msg:  "success",
		Ts:   time.Now().Unix(),
	}
	respJson, e := json.Marshal(resp)
	if e != nil {
		h.log.Errorf("json.Marshal: %v", e)
		return nil, e
	}
	var simpleCodec SimpleCodec = SimpleCodec{
		CurrentOrganize: orgId,
		CommandCode:     CommandAuthResult,
		Data:            respJson,
	}
	// 存储客户端连接
	h.clientUc.AddClient(c.RemoteAddr(), c, h.ifacePort.GetIface(simpleCodec.CurrentOrganize), clientVirtualIP.String())
	go func() {
		// 从interface读取数据并发送到客户端
		iface := h.orgUc.GetOrgInterface(orgId)
		if iface == nil {
			h.log.Errorf("GetOrgInterface: %v", ErrInvalidOrganize)
			return
		}
		client := h.clientUc.GetClient(clientVirtualIP.String())
		for {
			n, err := client.Buf.ReadFrom(iface)
			h.log.Debugf("ReadFrom: %d", n)
			if err != nil {
				h.log.Errorf("iface.Read: %v", err)
				continue
			}
			_, err = client.Buf.WriteTo(c)
			if err != nil {
				h.log.Errorf("c.Write: %v", err)
				continue
			}
		}
	}()
	return simpleCodec.Encode()
}

func (h *HandleUseCase) HandleCommandData(orgId uint16, data []byte) error {
	iface := h.orgUc.GetOrgInterface(orgId)
	if iface == nil {
		return ErrInvalidOrganize
	}
	_, err := iface.Write(data)
	if err != nil {
		h.log.Errorf("iface.Write: %v", err)
		return err
	}
	return nil
}

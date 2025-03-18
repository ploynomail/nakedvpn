package biz

import (
	"net"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/panjf2000/gnet/v2"
	"github.com/songgao/water"
)

type EnforcementClient struct {
	// The client's virtual IP address.
	ClientIP string `json:"client_ip"`
	// The client's virtual gw IP address.
	VirtualGateway string `json:"virtual_gateway"`
}

// Socket holds state about our connection.
type Client struct {
	clientIP  net.Addr
	conn      gnet.Conn
	iface     *water.Interface
	virtualIP string
}

type ClientUseCase struct {
	ClientHarbor map[string]*Client
	log          *log.Helper
}

func NewClientUseCase(logger log.Logger) *ClientUseCase {
	return &ClientUseCase{
		ClientHarbor: make(map[string]*Client),
		log:          log.NewHelper(log.With(logger, "module", "biz/client")),
	}
}

func (c *ClientUseCase) AddClient(clientIP net.Addr, conn gnet.Conn, iface *water.Interface, virtualIP string) {
	c.ClientHarbor[virtualIP] = &Client{
		clientIP:  clientIP,
		conn:      conn,
		iface:     iface,
		virtualIP: virtualIP,
	}
}

func (c *ClientUseCase) GetClient(virtualIP string) *Client {
	return c.ClientHarbor[virtualIP]
}

func (c *ClientUseCase) RemoveClient(virtualIP string) {
	delete(c.ClientHarbor, virtualIP)
}

func (c *ClientUseCase) CloseClient(virtualIP string) {
	client := c.GetClient(virtualIP)
	if client == nil {
		c.log.Errorf("CloseClient: client not found")
		return
	}
	client.conn.Close()
	client.iface.Close()
	c.RemoveClient(virtualIP)
}

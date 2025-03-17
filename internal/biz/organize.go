package biz

import (
	"NakedVPN/internal/utils"
	"net"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/patrickmn/go-cache"
	"github.com/songgao/water"
)

// 网络协议类型枚举
type NetworkProtocol uint8

const (
	NakedVPNProtocol NetworkProtocol = iota + 1
)

// 配额配置
type QuotaConfig struct {
	MaxUsers    int `json:"max_users"`    // 最大用户数
	MaxDevices  int `json:"max_devices"`  // 最大设备数
	BandwidthMB int `json:"bandwidth_mb"` // 带宽限制（MB/s）
}

// Organize 组织实体
// The maximum number of organizations is 65535
type Organize struct {
	ID             uint16          `json:"id" gorm:"primary_key;AUTO_INCREMENT;column:id;comment'ID'"`
	Name           string          `json:"name" gorm:"column:name;type:varchar(100);index:idx_username,unique;comment:'名称'"` // 组织名称
	ParentID       int64           `json:"parent_id" gorm:"column:parent_id;type:bigint;index:idx_parent_id;comment:'父级ID'"` // 父级ID
	Level          int32           `json:"level" gorm:"column:level;type:int;comment:'层级'"`                                  // 组织层级
	Children       []*Organize     `json:"children" gorm:"foreignkey:ID"`                                                    // 子组织列表
	SubnetCIDR     string          `json:"subnet_cidr" gorm:"column:subnet_cidr;type:varchar(20);comment:'子网CIDR'"`          // 子网CIDR（如 "10.1.0.0/24"）
	Gateway        string          `json:"gateway" gorm:"column:gateway;type:varchar(20);comment:'子网GW'"`                    // 子网网关（如 "10.1.0.1"）
	DNSServers     []string        `json:"dns_servers" gorm:"column:ds;type:varchar(20);comment:'子网DNS'"`                    // 租户专属DNS（如 ["8.8.8.8"]）
	Status         bool            `json:"status" gorm:"column:status;type:tinyint;comment:'状态'"`                            // 状态（启用/禁用）
	Protocol       NetworkProtocol `json:"protocol" gorm:"column:protocol;type:tinyint;comment:'网络协议'"`                      // 网络协议
	Quotas         *QuotaConfig    `json:"quotas" gorm:"column:quotas;type:json;comment:'配额配置'"`                             // 配额配置
	AdminUsers     StringArr       `json:"admin_users" gorm:"column:admin_users;type:json;comment:'管理员用户列表'"`                // 管理员用户列表
	CreatedAt      time.Time       `json:"created_at"`                                                                       // 创建时间
	UpdatedAt      time.Time       `json:"updated_at"`                                                                       // 更新时间
	AccessKey      string          `json:"access_key" gorm:"column:access_key;type:varchar(100);comment:'访问密钥'"`             // 访问密钥
	AdvancedConfig *MapAny         `json:"advanced_config" gorm:"column:advanced_config;type:json;comment:'高级网络配置'"`         // 高级网络配置(如防火墙规则、路由策略）
	virtualGateway string          `gorm:"-"`
	ipMap          []byte          `gorm:"-"`
	ipCount        int             `gorm:"-"`
	ipStart        net.IP          `gorm:"-"`
	ipEnd          net.IP          `gorm:"-"`
	mu             sync.Mutex      `gorm:"-"`
}

type OrganizeRepo interface {
	GetAllOrganizes() ([]*Organize, error)
}

type OrganizeUseCase struct {
	cache *cache.Cache
	repo  OrganizeRepo
	iPort *IfacePort
	log   *log.Helper
}

func NewOrganizeUseCase(repo OrganizeRepo, iPort *IfacePort, logger log.Logger) *OrganizeUseCase {
	l := log.NewHelper(log.With(logger, "module", "biz/organize"))
	ca := cache.New(cache.NoExpiration, 0)
	orgs, err := repo.GetAllOrganizes()
	if err != nil {
		l.Errorf("failed to get all organizes: %v", err)
	}
	for _, org := range orgs {
		_, ipNet, err := net.ParseCIDR(org.SubnetCIDR)
		if err != nil {
			l.Errorf("failed to parse CIDR: %v", err)
			continue
		}

		// 计算 IP 范围
		ipStart := ipNet.IP.To4()
		if ipStart == nil {
			l.Errorf("failed to get IP: %v", err)
			continue
		}
		mask := ipNet.Mask
		ipEnd := make(net.IP, len(ipStart))
		copy(ipEnd, ipStart)
		for i := 0; i < len(mask); i++ {
			ipEnd[i] |= ^mask[i]
		}

		// 计算 IP 数量
		ipCount := 0
		for i := range ipStart {
			ipCount = ipCount*256 + int(ipStart[i])
		}
		ipCount = int(ipEnd[len(ipEnd)-1]) - int(ipStart[len(ipStart)-1]) + 1
		//去除网关和广播地址
		ipCount -= 2
		ipStart[len(ipStart)-1]++
		ipEnd[len(ipEnd)-1]--
		org.virtualGateway = ipEnd.String()
		ipEnd[len(ipEnd)-1]--
		// 初始化ip位图
		bitmapSize := (ipCount + 7) / 8
		org.ipMap = make([]byte, bitmapSize)
		org.ipCount = ipCount
		org.ipStart = ipStart
		org.ipEnd = ipEnd
		ca.Set(strconv.Itoa(int(org.ID)), org, cache.NoExpiration)
	}

	return &OrganizeUseCase{
		cache: ca,
		repo:  repo,
		iPort: iPort,
		log:   l,
	}
}

func (uc *OrganizeUseCase) AuthAccessKey(id uint16, accessKey string) bool {
	org, ok := uc.cache.Get(strconv.Itoa(int(id)))
	if !ok {
		return false
	}
	return org.(*Organize).AccessKey == accessKey
}

func (uc *OrganizeUseCase) GetOrgInterface(id uint16) *water.Interface {
	return uc.iPort.GetIface(id)
}

func (uc *OrganizeUseCase) PrepareTun(id uint16) error {
	org, ok := uc.cache.Get(strconv.Itoa(int(id)))
	if !ok {
		return ErrInvalidOrganize
	}
	tunDevName := org.(*Organize).Name
	// Check if the tun device already exists

	if i := uc.iPort.GetIface(id); i != nil {
		uc.log.Infof("tun device already exists: %v", tunDevName)
		return nil
	} else {
		var iface *water.Interface
		iface, err := water.New(water.Config{
			PlatformSpecificParams: water.PlatformSpecificParams{
				Name: tunDevName,
			},
			DeviceType: water.TUN,
		})
		if err != nil {
			uc.log.Errorf("failed to create tun device: %v", err)
			return err
		}

		// 设置接口IP地址
		// TODO: 处理错误
		lip, _ := uc.GetOrganizeServerIp(id)
		cmd := exec.Command("ip", "addr", "add", lip, "dev", iface.Name())
		if err := cmd.Run(); err != nil {
			uc.log.Errorf("Failed to set IP address for interface: %v", err)
			return err
		}

		// 设置接口为UP状态
		cmd = exec.Command("ip", "link", "set", iface.Name(), "up")
		if err := cmd.Run(); err != nil {
			uc.log.Errorf("Failed to set interface UP: %v", err)
			return err
		}
		uc.iPort.SetIface(id, iface)
		uc.log.Infof("tun device created: %v", iface.Name())
		return nil
	}

}

func (uc *OrganizeUseCase) GetOrganizeServerIp(id uint16) (string, error) {
	org, ok := uc.cache.Get(strconv.Itoa(int(id)))
	if !ok {
		return "", ErrInvalidOrganize
	}
	orgInfo := org.(*Organize)
	lip, err := utils.GetLastIP(orgInfo.SubnetCIDR)
	if err != nil {
		uc.log.Errorf("failed to get last IP: %v", err)
		return "", err
	}
	return lip, nil
}

func (uc *OrganizeUseCase) GetOrgVirtualGateway(id uint16) (string, error) {
	org, ok := uc.cache.Get(strconv.Itoa(int(id)))
	if !ok {
		return "", ErrInvalidOrganize
	}
	return org.(*Organize).virtualGateway, nil
}

func (uc *OrganizeUseCase) AllocateIpForOrg(id uint16) (net.IP, error) {
	org, ok := uc.cache.Get(strconv.Itoa(int(id)))
	if !ok {
		return nil, ErrInvalidOrganize
	}
	orgInfo := org.(*Organize)
	orgInfo.mu.Lock()
	defer orgInfo.mu.Unlock()

	for i := 0; i < orgInfo.ipCount; i++ {
		byteIndex := i / 8
		bitIndex := i % 8

		if (orgInfo.ipMap[byteIndex] & (1 << bitIndex)) == 0 {
			// 标记为已分配
			orgInfo.ipMap[byteIndex] |= 1 << bitIndex

			// 计算 IP 地址
			ip := make(net.IP, len(orgInfo.ipStart))
			// 从起始 IP + 1 开始
			copy(ip, orgInfo.ipStart)
			for j := len(ip) - 1; j >= 0; j-- {
				ip[j] += byte(i)
				if ip[j] != 0 {
					break
				}
			}

			return ip, nil
		}
	}
	return nil, ErrNoAvailableIp
}

func (uc *OrganizeUseCase) Release(id uint16, ip net.IP) error {
	org, ok := uc.cache.Get(strconv.Itoa(int(id)))
	if !ok {
		return ErrInvalidOrganize
	}
	orgInfo := org.(*Organize)
	orgInfo.mu.Lock()
	defer orgInfo.mu.Unlock()

	// 检查 IP 是否在网段内
	_, orgCidr, _ := net.ParseCIDR(orgInfo.SubnetCIDR)
	if !orgCidr.Contains(ip) {
		return ErrInvalidIp
	}

	// 计算 IP 的偏移量
	offset := 0
	for i := range ip {
		offset = offset*256 + int(ip[i]-orgInfo.ipStart[i])
	}

	// 检查偏移量是否有效
	if offset < 0 || offset >= orgInfo.ipCount {
		return ErrInvalidIp
	}

	// 标记为未分配
	byteIndex := offset / 8
	bitIndex := offset % 8
	orgInfo.ipMap[byteIndex] &^= 1 << bitIndex

	return nil
}

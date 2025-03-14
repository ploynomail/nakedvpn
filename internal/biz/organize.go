package biz

import (
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/patrickmn/go-cache"
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
}

type OrganizeRepo interface {
	GetAllOrganizes() ([]*Organize, error)
}

type OrganizeUseCase struct {
	cache *cache.Cache
	repo  OrganizeRepo
	log   *log.Helper
}

func NewOrganizeUseCase(repo OrganizeRepo, logger log.Logger) *OrganizeUseCase {
	l := log.NewHelper(log.With(logger, "module", "biz/organize"))
	ca := cache.New(cache.NoExpiration, 0)
	orgs, err := repo.GetAllOrganizes()
	if err != nil {
		l.Errorf("failed to get all organizes: %v", err)
	}
	for _, org := range orgs {
		ca.Set(strconv.Itoa(int(org.ID)), org, cache.NoExpiration)
	}
	return &OrganizeUseCase{
		cache: ca,
		repo:  repo,
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

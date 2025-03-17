package biz

import (
	"net"
	"strconv"
	"testing"

	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
)

func TestAllocateIpForOrg(t *testing.T) {
	// Mock data
	org := &Organize{
		ID:         1,
		SubnetCIDR: "10.1.0.0/24",
		ipMap:      make([]byte, 32),
	}

	ca := cache.New(cache.NoExpiration, 0)

	_, ipNet, err := net.ParseCIDR(org.SubnetCIDR)
	if err != nil {
		t.Errorf("failed to parse CIDR: %v", err)
	}

	// 计算 IP 范围
	ipStart := ipNet.IP.To4()
	if ipStart == nil {
		t.Errorf("failed to get IP: %v", err)
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
	ipEnd[len(ipEnd)-1]--
	// 初始化ip位图
	bitmapSize := (ipCount + 7) / 8
	org.ipMap = make([]byte, bitmapSize)
	org.ipCount = ipCount
	org.ipStart = ipStart
	org.ipEnd = ipEnd

	ca.Set(strconv.Itoa(int(org.ID)), org, cache.NoExpiration)

	uc := &OrganizeUseCase{
		cache: ca,
	}

	// Test case 1: Allocate first IP
	ip, err := uc.AllocateIpForOrg(org.ID)
	assert.NoError(t, err)
	assert.Equal(t, net.IPv4(10, 1, 0, 1).To4(), ip)

	// Test case 2: Allocate second IP
	ip, err = uc.AllocateIpForOrg(org.ID)
	assert.NoError(t, err)
	assert.Equal(t, net.IPv4(10, 1, 0, 2).To4(), ip)

	// Test case 3: Allocate all IPs
	for i := 2; i < org.ipCount; i++ {
		_, err = uc.AllocateIpForOrg(org.ID)
		assert.NoError(t, err)
	}

	// Test case 4: No available IPs
	ip, err = uc.AllocateIpForOrg(org.ID)
	assert.Error(t, err)
	assert.Nil(t, ip)
	assert.Equal(t, ErrNoAvailableIp, err)
}

func TestAllocateIpForOrg_InvalidOrganize(t *testing.T) {
	ca := cache.New(cache.NoExpiration, 0)
	uc := &OrganizeUseCase{
		cache: ca,
	}

	// Test case: Invalid organize ID
	ip, err := uc.AllocateIpForOrg(999)
	assert.Error(t, err)
	assert.Nil(t, ip)
	assert.Equal(t, ErrInvalidOrganize, err)
}

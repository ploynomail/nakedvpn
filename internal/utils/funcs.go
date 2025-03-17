package utils

import (
	"fmt"
	"net"
)

// 解析CIDR地址
func ParseCIDR(cidr string) (string, string, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", "", err
	}
	return ipnet.IP.String(), ipnet.Mask.String(), nil
}

// 获取CIDR最后一个可用的IP地址
func GetLastIP(cidr string) (string, error) {
	// 解析CIDR地址

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", fmt.Errorf("解析CIDR地址失败: %w", err)
	}

	// 获取网络的IP地址和掩码
	ip := ipNet.IP
	mask := ipNet.Mask

	// 将IP地址转换为32位整数
	ipUint := ip.To4()
	if ipUint == nil {
		return "", fmt.Errorf("不支持的IP地址格式")
	}

	// 计算网络地址
	networkAddr := make(net.IP, len(ip))
	copy(networkAddr, ip)
	for i := 0; i < len(mask); i++ {
		networkAddr[i] &= mask[i]
	}

	// 计算广播地址（即最后一个IP地址）
	broadcastAddr := make(net.IP, len(ip))
	copy(broadcastAddr, networkAddr)
	for i := 0; i < len(mask); i++ {
		broadcastAddr[i] |= ^mask[i]
	}
	// 广播地址减1即为最后一个IP地址
	broadcastAddr[len(broadcastAddr)-1]--
	// 将广播地址转换为字符串
	return broadcastAddr.String(), nil
}

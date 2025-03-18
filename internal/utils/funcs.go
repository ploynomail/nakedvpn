package utils

import (
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
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

func AnalyzePacket(packet gopacket.Packet) {
	// 打印数据包的以太网头部信息
	ethLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethLayer != nil {
		eth, _ := ethLayer.(*layers.Ethernet)
		fmt.Printf("以太网头部: 源MAC=%s, 目的MAC=%s\n", eth.SrcMAC, eth.DstMAC)
	}

	// 打印数据包的 IP 头部信息
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)
		fmt.Printf("IP 头部: 源IP=%s, 目的IP=%s\n", ip.SrcIP, ip.DstIP)
	}

	// 打印数据包的 ICMP 头部信息
	icmpLayer := packet.Layer(layers.LayerTypeICMPv4)
	if icmpLayer != nil {
		icmp, _ := icmpLayer.(*layers.ICMPv4)
		fmt.Printf("ICMP 头部: 类型=%d, 代码=%d, 校验和=%d\n", icmp.TypeCode.Type(), icmp.TypeCode.Code(), icmp.Checksum)
		fmt.Printf("ICMP 负载: %v\n", icmp.Payload)
	}

	// 打印数据包的 TCP 头部信息
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		tcp, _ := tcpLayer.(*layers.TCP)
		fmt.Printf("TCP 头部: 源端口=%d, 目的端口=%d, 序列号=%d, 确认号=%d\n", tcp.SrcPort, tcp.DstPort, tcp.Seq, tcp.Ack)
	}

	// 打印数据包的 UDP 头部信息
	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer != nil {
		udp, _ := udpLayer.(*layers.UDP)
		fmt.Printf("UDP 头部: 源端口=%d, 目的端口=%d\n", udp.SrcPort, udp.DstPort)
	}

	// 打印其他信息（如未识别的协议）
	if packet.NetworkLayer() == nil {
		fmt.Println("未识别的网络层协议")
	} else if packet.TransportLayer() == nil {
		fmt.Println("未识别的传输层协议")
	} else if packet.ApplicationLayer() == nil {
		fmt.Println("未识别的应用层协议")
	} else {
		fmt.Printf("应用层负载: %s\n", packet.ApplicationLayer().Payload())
	}

	fmt.Println("--------------------------")
}

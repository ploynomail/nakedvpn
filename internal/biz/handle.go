package biz

type Command uint16

const (
	// CommandHeartbeat 心跳包
	CommandHeartbeat Command = iota + 1
	// CommandReqAuth 请求认证
	CommandReqAuth
	// CommandAuth 认证包
	CommandAuth
	// CommandData 数据包
	CommandData
	// CommandClose 关闭连接
	CommandClose
)

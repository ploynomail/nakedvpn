package biz

import (
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewOrganizeUseCase,
	NewHandleUseCase,
	NewHandleClientUseCase,
	NewIfacePort,
	NewClientUseCase,
)

var (
	ErrUnknown          = errors.New(10001, "未知错误", "未知错误")
	ErrIncompletePacket = errors.New(10002, "incomplete packet", "incomplete packet")
	ErrNotFound         = errors.New(10003, "未找到", "未找到")
	ErrAuth             = errors.New(10004, "认证失败", "认证失败")
	ErrInvalidOrganize  = errors.New(10005, "无效的组织", "无效的组织")
	ErrInvalidData      = errors.New(10006, "无效的数据", "无效的数据")
	ErrAuthFailed       = errors.New(10007, "认证失败", "认证失败")
	ErrNoAvailableIp    = errors.New(10008, "没有可用的IP", "没有可用的IP")
	ErrInvalidIp        = errors.New(10009, "无效的IP", "无效的IP")
)

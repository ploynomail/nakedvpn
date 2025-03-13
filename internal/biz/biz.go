package biz

import (
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewOrganizeUseCase)

var (
	ErrUnknown          = errors.New(10001, "未知错误", "未知错误")
	ErrIncompletePacket = errors.New(10002, "incomplete packet", "incomplete packet")
	ErrNotFound         = errors.New(10003, "未找到", "未找到")
	ErrAuth             = errors.New(10004, "认证失败", "认证失败")
)

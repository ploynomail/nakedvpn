package biz

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/panjf2000/gnet/v2"
	"github.com/valyala/bytebufferpool"
)

const (
	organizeNumberSize = 2                                        // 组织号字节长度 (2字节)
	commandCodeSize    = 2                                        // 命令码字节长度 (2字节)
	headerSize         = organizeNumberSize + commandCodeSize + 4 // 完整头部长度 (8字节)
)

// 协议结构：
//
// 0           2           4           8          (字节位置)
// +-----------+-----------+-----------+
// |OrganizeID| CommandCode| BodyLen  |
// +-----------+-----------+-----------+
// |                                   |
// +                                   +
// |           body bytes              |
// +                                   +
// |            ... ...                |
// +-----------------------------------+
//
// 说明：
// 1. OrganizeID：组织号（2字节）
// 2. CommandCode：命令码（2字节）
// 3. BodyLen：Body长度（4字节）
// 4. Body：数据部分

type SimpleCodec struct {
	CurrentOrganize uint16  // 当前组织标识（编码时使用）
	CommandCode     Command // 当前命令码（编码时使用）
	Data            []byte  // 数据
}

func (codec SimpleCodec) Encode() ([]byte, error) {
	header := make([]byte, headerSize)

	// 写入组织号（大端序）
	binary.BigEndian.PutUint16(header[0:2], codec.CurrentOrganize)

	// 写入命令码（大端序）
	commandCode := uint16(codec.CommandCode) // 替换为实际命令码
	binary.BigEndian.PutUint16(header[2:4], commandCode)

	// 写入Body长度
	bodyLen := uint32(len(codec.Data))
	binary.BigEndian.PutUint32(header[4:8], bodyLen)

	return append(header, codec.Data...), nil
}

func (codec *SimpleCodec) Decode(c gnet.Conn) error {
	// 读取完整头部
	Buf := bytebufferpool.Get()
	defer bytebufferpool.Put(Buf)
	n, err := io.CopyN(Buf, c, headerSize)
	if err != nil || n < headerSize {
		if errors.Is(err, io.ErrShortBuffer) {
			return ErrIncompletePacket
		}
		return err
	}

	// 解析组织号
	organizeID := binary.BigEndian.Uint16(Buf.B[0:2])

	// 解析命令码
	commandCode := binary.BigEndian.Uint16(Buf.B[2:4])
	// 解析Body长度
	bodyLen := binary.BigEndian.Uint32(Buf.B[4:8])

	// 检查数据长度
	if c.InboundBuffered() < int(bodyLen) {
		fmt.Println("c.InboundBuffered() < int(bodyLen)", c.InboundBuffered(), int(bodyLen))
		return ErrIncompletePacket
	}

	// 校验组织合法性
	if err := validateOrganization(organizeID); err != nil {
		return err
	}

	// 读取完整数据包 TODO: 优化读取方式 Zero-Copy
	body := bytebufferpool.Get()
	defer bytebufferpool.Put(body)
	n, err = io.CopyN(body, c, int64(bodyLen))
	if err != nil || n < int64(bodyLen) {
		if errors.Is(err, io.ErrShortBuffer) {
			return ErrIncompletePacket
		}
	}

	// 更新Codec
	codec.CurrentOrganize = organizeID
	codec.CommandCode = Command(commandCode)
	codec.Data = body.B
	return nil
}

// 协议解包(直接获取Body)
func Unpack(c net.Conn) (Command, []byte, error) {
	Buf := bytebufferpool.Get()
	defer bytebufferpool.Put(Buf)
	_, err := io.CopyN(Buf, c, 8)
	if err != nil {
		if errors.Is(err, io.ErrShortBuffer) {
			return 0, nil, err
		}
		return 0, nil, err
	}
	bodysize := binary.BigEndian.Uint32(Buf.B[4:8])
	body := bytebufferpool.Get()
	defer bytebufferpool.Put(body)
	_, err = io.CopyN(body, c, int64(bodysize))
	if err != nil {
		if errors.Is(err, io.ErrShortBuffer) {
			return 0, nil, err
		}
		return 0, nil, err
	}
	command := binary.BigEndian.Uint16(Buf.B[2:4])
	return Command(command), body.B, nil
}

// 示例验证逻辑（需实现具体业务规则）
func validateOrganization(organizeID uint16) error {
	return nil
}

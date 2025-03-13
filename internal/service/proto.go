package service

import (
	"NakedVPN/internal/biz"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/panjf2000/gnet/v2"
	"github.com/valyala/bytebufferpool"
)

const (
	organizeNumberSize = 2                                        // 组织号字节长度
	commandCodeSize    = 2                                        // 命令码字节长度
	headerSize         = organizeNumberSize + commandCodeSize + 4 // 完整头部长度
)

// 修改后的协议结构：
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
type SimpleCodec struct {
	CurrentOrganize uint16 // 当前组织标识（编码时使用）
	CommandCode     uint16 // 当前命令码（编码时使用）
	Data            []byte // 数据
}

func (codec SimpleCodec) Encode() ([]byte, error) {
	header := make([]byte, headerSize)

	// 写入组织号（大端序）
	binary.BigEndian.PutUint16(header[0:2], codec.CurrentOrganize)

	// 写入命令码（示例值，根据业务需求定义）
	commandCode := uint16(codec.CommandCode) // 替换为实际命令码
	binary.BigEndian.PutUint16(header[2:4], commandCode)

	// 写入Body长度
	bodyLen := uint32(len(codec.Data))
	binary.BigEndian.PutUint32(header[4:8], bodyLen)

	return append(header, codec.Data...), nil
}

func (codec *SimpleCodec) Decode(c gnet.Conn) error {
	// 第一步：读取完整头部
	Buf := bytebufferpool.Get()
	defer bytebufferpool.Put(Buf)
	n, err := Buf.ReadFrom(c)
	if err != nil || n < headerSize {
		fmt.Println("read from conn error:", err)
		if errors.Is(err, io.ErrShortBuffer) {
			return biz.ErrIncompletePacket
		}
		return err
	}

	// 第二步：解析组织号
	organizeID := binary.BigEndian.Uint16(Buf.B[0:2])

	// 第三步：解析命令码
	commandCode := binary.BigEndian.Uint16(Buf.B[2:4])

	// 第四步：解析Body长度
	bodyLen := binary.BigEndian.Uint32(Buf.B[4:8])
	totalLen := headerSize + int(bodyLen)

	// 第五步：检查数据完整性
	if c.InboundBuffered() < totalLen {
		return biz.ErrIncompletePacket
	}

	// 第六步：提取组织上下文（示例实现）
	if err := validateOrganization(organizeID); err != nil {
		return err
	}

	// 第七步：读取完整数据包
	fullPacket, _ := c.Peek(totalLen)
	body := make([]byte, bodyLen)
	copy(body, fullPacket[headerSize:totalLen])

	// 第八步：丢弃已处理字节
	_, _ = c.Discard(totalLen)
	codec.CurrentOrganize = organizeID
	codec.CommandCode = commandCode
	codec.Data = body
	return nil
}

func (codec *SimpleCodec) DecodeForStdNet(c net.Conn) error {
	// 第一步：读取完整头部
	Buf := bytebufferpool.Get()
	defer bytebufferpool.Put(Buf)
	n, err := io.CopyN(Buf, c, headerSize)
	fmt.Println("read from conn:", n)
	if err != nil || n < headerSize {
		fmt.Println("read from conn error:", err)
		if errors.Is(err, io.ErrShortBuffer) {
			return biz.ErrIncompletePacket
		}
		return err
	}

	// 第二步：解析组织号
	organizeID := binary.BigEndian.Uint16(Buf.B[0:2])

	// 第三步：解析命令码
	commandCode := binary.BigEndian.Uint16(Buf.B[2:4])

	// 第四步：解析Body长度
	bodyLen := binary.BigEndian.Uint32(Buf.B[4:8])

	body := bytebufferpool.Get()
	defer bytebufferpool.Put(body)
	n, err = io.CopyN(body, c, int64(bodyLen))
	if err != nil || n < int64(bodyLen) {
		fmt.Println("read from conn error:", err)
		if errors.Is(err, io.ErrShortBuffer) {
			return biz.ErrIncompletePacket
		}
	}
	codec.CurrentOrganize = organizeID
	codec.CommandCode = commandCode
	codec.Data = body.B
	return nil
}

// 协议解包（适配不同场景）
func (codec SimpleCodec) Unpack(buf []byte) ([]byte, error) {
	if len(buf) < headerSize {
		return nil, biz.ErrIncompletePacket
	}

	// 校验组织合法性
	organizeID := binary.BigEndian.Uint16(buf[0:2])
	if err := validateOrganization(organizeID); err != nil {
		return nil, err
	}

	// 获取实际Body长度
	bodyLen := binary.BigEndian.Uint32(buf[4:8])
	totalLen := headerSize + int(bodyLen)
	if len(buf) < totalLen {
		return nil, biz.ErrIncompletePacket
	}

	return buf[:totalLen], nil
}

// 示例验证逻辑（需实现具体业务规则）
func validateOrganization(organizeID uint16) error {
	return nil
}

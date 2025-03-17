package biz

import (
	"sync"

	"github.com/songgao/water"
)

var once sync.Once
var IfacePortInstance *IfacePort

type IfacePort struct {
	Imap map[uint16]*water.Interface
	sync.Mutex
}

func NewIfacePort() *IfacePort {
	// Initialize the map signle instance
	once.Do(func() {
		IfacePortInstance = &IfacePort{
			Imap: make(map[uint16]*water.Interface),
		}
		// 初始化单例对象的状态
	})
	return IfacePortInstance
}

func (i *IfacePort) GetIface(name uint16) *water.Interface {
	i.Lock()
	defer i.Unlock()
	if iface, ok := i.Imap[name]; ok {
		return iface
	}
	return nil
}

func (i *IfacePort) SetIface(name uint16, iface *water.Interface) {
	i.Lock()
	defer i.Unlock()
	i.Imap[name] = iface
}

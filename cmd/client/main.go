package main

import (
	"NakedVPN/internal/biz"
	"NakedVPN/internal/service"
	"flag"
	"fmt"
	"net"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/valyala/bytebufferpool"
)

func logErr(err error) {
	if err != nil {
		log.Errorf("error: %v", err)
		panic(err)
	}
}

func main() {
	var (
		network string
		addr    string
	)

	// Example command: go run main.go --network tcp --address ":9000" --concurrency 100 --packet_size 1024 --packet_batch 20 --packet_count 1000
	flag.StringVar(&network, "network", "tcp", "--network tcp")
	flag.StringVar(&addr, "address", "127.0.0.1:9001", "--address 127.0.0.1:9001")
	flag.Parse()

	log.Infof("start clients...")
	runClient(network, addr)
	log.Infof("clients are done")
}

func runClient(network, addr string) {
	c, err := net.Dial(network, addr)
	logErr(err)
	log.Infof("connection=%s starts...", c.LocalAddr().String())
	defer func() {
		log.Infof("connection=%s stops...", c.LocalAddr().String())
		c.Close()
	}()
	var codec service.SimpleCodec = service.SimpleCodec{}
	err = codec.DecodeForStdNet(c)
	logErr(err)
	if codec.CommandCode == uint16(biz.CommandReqAuth) {
		fmt.Println(string(codec.Data))
	}
}

func Init(c net.Conn) {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)
	buf.ReadFrom(c)

	var codec service.SimpleCodec = service.SimpleCodec{}
	f, err := codec.Unpack(buf.B)
	logErr(err)

	log.Infof("response: %s", string(f))
}

package main

import (
	"NakedVPN/internal/conf"
	"NakedVPN/internal/server"
	"flag"
	"os"

	kzap "github.com/go-kratos/kratos/contrib/log/zap/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, ns *server.NetClient) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			ns,
		),
	)
}

func main() {
	flag.Parse()
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()
	// load config
	if err := c.Load(); err != nil {
		panic(err)
	}
	// init log
	var bc *conf.Client = &conf.Client{}
	if err := c.Scan(bc); err != nil {
		panic(err)
	}
	// override version
	bc.Version = Version
	// init logger
	logLevel, err := zap.ParseAtomicLevel(bc.Logger.Level)
	if err != nil {
		panic(err)
	}
	// init logger: initial fields
	var initFields map[string]interface{}
	if bc.Logger.InitialFields != nil {
		initFields = make(map[string]interface{})
		for k, v := range bc.Logger.InitialFields {
			initFields[k] = v
		}
	}
	// zap config: https://pkg.go.dev/go.uber.org/zap#Config
	cfg, err := zap.Config{
		Level:            logLevel,
		Encoding:         bc.Logger.Encoding,
		OutputPaths:      bc.Logger.Path,
		ErrorOutputPaths: bc.Logger.ErrorPath,
		InitialFields:    initFields,
	}.Build()
	if err != nil {
		panic(err)
	}
	zaplog := kzap.NewLogger(cfg)
	defer func() { _ = zaplog.Sync() }()
	logger := log.With(zaplog,
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
	)
	// init server
	app, cleanup, err := wireApp(bc, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()
	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}

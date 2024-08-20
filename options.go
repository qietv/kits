package kits

import (
	"fmt"
	"github.com/qietv/kits/discovery"
	"github.com/qietv/kits/metrics"
	"github.com/qietv/kits/utils"
	"github.com/qietv/qgrpc"
	"google.golang.org/grpc"
	"net"
	"os"
	"strings"
)

type QGrpc struct {
	Conf             *qgrpc.Config
	GrpcRegisterFunc func(grpcServer *grpc.Server)
	Options          []grpc.ServerOption
}
type BuildInfo struct {
	GoVersion  string
	AppVersion string
	GitVersion string
	BuildTime  string
}
type Environment int

const (
	Develop Environment = iota
	Test
	Canary
	Preview
	Online
)

func (e Environment) String() string {
	switch e {
	case Develop:
		return "dev"
	case Test:
		return "test"
	case Canary:
		return "canary"
	case Preview:
		return "pre"
	case Online:
		return "online"
	}
	return "unknown"
}

func (e *Environment) Set(env string) error {
	switch strings.ToLower(env) {
	case "dev", "develop":
		*e = Develop
	case "test":
		*e = Test
	case "canary":
		*e = Canary
	case "pre":
		*e = Preview
	case "online", "prod":
		*e = Online
	case "":
		*e = Develop
	default:
		return fmt.Errorf("Not Support (%s)", env)
	}
	return nil
}

type options struct {
	Metric          *metrics.Metric
	consul          *discovery.Consul
	debug           bool
	id              string
	name            string
	host            string
	port            int
	build           *BuildInfo
	Env             Environment
	Grpc            *QGrpc
	GrpcInterecpter *grpc.ServerOption
	shutdownFunc    func(s os.Signal) error
	logger          utils.Logger
}

var defaultOptions = options{
	name: "qietv",
	host: utils.GetIP(),
	port: 8808,
	Env:  Develop,
}

type Option interface {
	apply(*options)
}
type funcOption struct {
	f func(*options)
}

func (f *funcOption) apply(opts *options) {
	f.f(opts)
}
func newFuncOption(f func(*options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

// Metrics server metrics Info
func Metrics(opts *metrics.Metric) Option {
	return newFuncOption(func(o *options) {
		if opts == nil {
			return
		}
		o.Metric = metrics.DefaultMetrics
		o.Metric.Collectors = opts.Collectors
		if opts.Path != "" {
			o.Metric.Path = opts.Path
		}
		if opts.Address != "" {
			o.Metric.Address = opts.Address
		}
		o.Metric.Handler = opts.Handler
		if opts.Port != 0 {
			o.Metric.Port = opts.Port
		} else if opts.Listen != "" {
			o.Metric.Listen = opts.Listen
			if tcpAddrs, e := net.ResolveTCPAddr("tcp", opts.Listen); e == nil {
				o.Metric.Port = tcpAddrs.Port
			}
		}
		o.Metric.NodeName = opts.NodeName
	})
}

// Grpc server grpc server Info
func Grpc(opts *QGrpc) Option {
	return newFuncOption(func(o *options) {
		o.Grpc = opts
	})
}

// Consul server Consul discovery Info
func Consul(opts *discovery.Consul) Option {
	return newFuncOption(func(o *options) {
		o.consul = opts
	})
}

// Host server main transport ip
func Host(host string) Option {
	return newFuncOption(func(o *options) {
		o.host = host
	})
}

// Port server main transport port
func Port(port int) Option {
	return newFuncOption(func(o *options) {
		o.port = port
	})
}

// Name server name
func Name(name string) Option {
	return newFuncOption(func(o *options) {
		o.name = name
	})
}

// ServerID server id
func ServerID(id string) Option {
	return newFuncOption(func(o *options) {
		o.id = id
	})
}

// Build server build info
func Build(info *BuildInfo) Option {
	return newFuncOption(func(o *options) {
		o.build = info
	})
}

// Env server run environment
func Env(env Environment) Option {
	return newFuncOption(func(o *options) {
		o.Env = env
	})
}

// Debug server debug or false, debug will run pprof on http 9910
func Debug(debug bool) Option {
	return newFuncOption(func(o *options) {
		o.debug = debug
	})
}

// Logger server logger
func Logger(logger utils.Logger) Option {
	return newFuncOption(func(o *options) {
		o.logger = logger
	})
}
func ShutdownFunc(fn func(s os.Signal) error) Option {
	return newFuncOption(func(o *options) {
		o.shutdownFunc = fn
	})
}

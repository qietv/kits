package kits

import (
	"flag"
	"git.qietv.work/go-public/kits/discovery"
	"git.qietv.work/go-public/kits/metrics"
	"git.qietv.work/go-public/logkit"
	"git.qietv.work/go-public/qgrpc"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var (
	server     *Server
	opts       options
	serverId   string
	consulAddr string
	datacenter string
	env        Environment
)

func init() {
	flag.StringVar(&serverId, "sid", os.Getenv("SERVER_ID"), "server id also usage env `SERVER_ID`")
	if serverId == "" {
		serverId ,_ = os.Hostname()
	}
	flag.StringVar(&consulAddr, "consul.addr", os.Getenv("CONSUL_ADDR"), "consul register address also usage env `CONSUL_ADDR`")
	flag.StringVar(&datacenter, "datacenter", os.Getenv("DATACENTER"), "consul datacenter  also env `DATACENTER`")
	flag.Var(&env, "env", "server environment  default `dev` also usage env `ENV`")
	if env == 0 {
		err := env.Set(os.Getenv("ENV"))
		if err != nil {
			log.Fatal("env only support dev(develop)/test/pre/canary/online(prod)")
		}
	}
}

type Server struct {
	Grpc      *qgrpc.Server
	disCancel func()
}

func New(opt ...Option) (s *Server, err error) {
	s = &Server{}
	opts = defaultOptions
	for _, o := range opt {
		o.apply(&opts)
	}
	if opts.id == "" {
		opts.id = serverId
	}
	if opts.Env == 0 {
		opts.Env = env
	}
	if opts.consul == nil && consulAddr != "" && datacenter != ""{
		opts.consul = &discovery.Consul{
			Endpoint:   consulAddr,
			Datacenter: datacenter,
		}
	}
	if opts.Grpc != nil {
		if opts.Grpc.GrpcRegisterFunc == nil {
			panic("register must be set")
		}
		s.Grpc, err = qgrpc.New(opts.Grpc.Conf, opts.Grpc.GrpcRegisterFunc)
		if err != nil {
			return
		}
	}
	if opts.consul != nil {
		var (
			tags = []string{opts.Env.String()}
			meta = map[string]string{}
		)

		if opts.Metric != nil {
			tags = append(tags, "metrics")
			//meta["__meta_consul_service_id"] = opts.Metric.NodeName
			//meta["__meta_consul_address"] = "172.17.3.188:9909"
			s.disCancel = discovery.NewConsul(opts.consul.Endpoint, opts.consul.Datacenter, opts.id).RegisterConsul(opts.name, opts.host, opts.port, opts.Metric.Port, tags, meta)
		} else {
			s.disCancel = discovery.NewConsul(opts.consul.Endpoint, opts.consul.Datacenter, opts.id).RegisterConsul(opts.name, opts.host, opts.port, 0, tags, meta)
		}
	}
	if opts.Metric != nil {
		if opts.build != nil {
			prometheus.Register(prometheus.NewGauge(prometheus.GaugeOpts{
				Namespace:   "qietv",
				Name:        "build_info",
				Help:        "This mirco service build info",
				ConstLabels: map[string]string{"build": opts.build.BuildTime, "app_version": opts.build.AppVersion, "git_commit": opts.build.GitVersion, "go_version": runtime.Version()},
			}))
		}
		if qgrpc.Metrics != nil {
			metrics.Register(qgrpc.Metrics)
		}
		metrics.Register(opts.Metric.Collectors...)
		metrics.InitMetrics(opts.Metric)
	}

	server = s
	logkit.Infof("%s server start.", opts.name)
	ShutdownHook()
	return
}

func ShutdownHook() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			server.Grpc.GracefulStop()
			if server.disCancel != nil {
				server.disCancel()
			}
			logkit.Infof("%s-server exit {id: %s host:%s port:%d}", opts.name, opts.id, opts.host, opts.port)
			if opts.shutdownFunc != nil {
				opts.shutdownFunc(s)
			}
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

package kits

import (
	"git.qietv.work/go-public/kits/discovery"
	"git.qietv.work/go-public/kits/metrics"
	"git.qietv.work/go-public/logkit"
	"git.qietv.work/go-public/qgrpc"
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var (
	server *Server
	opts   options
)

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

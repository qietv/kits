#### Kits 
> Kits is a microframework for Qietv go developing

#### Usage
```golang
	kits.New(
        kits.Grpc(&kits.QGrpc{
            Conf:             &qgrpc.Config{
                Network:           "tcp",
                Addr:              ":8808",
                Interceptor:       nil,
            },
            GrpcRegisterFunc: func(grpcServer *grpc.Server) {
            //pb.RegisterGRPCLotteryServer(grpcServer, &lotteryServer{s})
            },
        }),
        kits.Name(conf.Name),
        kits.Metrics(&metrics.Metric{
            Port:       9909,
            Handler:    httpServer.Router,
        }),
        kits.Consul(&discovery.Consul{
            Endpoint:   "http://172.17.3.79:8500",
            Datacenter: "tx",
        }),
        kits.Port(8808),
        kits.ServerID(conf.HostName),
        kits.Build(&kits.BuildInfo{
            GoVersion:  runtime.Version(),
            AppVersion: conf.Version,
            GitVersion: conf.Build,
            BuildTime:  conf.Build,
        }),
        kits.ShutdownFunc(func(s os.Signal) error {
            logkit.Exit()
            return nil
        }), 
    )
```
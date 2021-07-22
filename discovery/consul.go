package discovery

import (
	"context"
	"fmt"
	log "git.qietv.work/go-public/logkit"
	consul "github.com/hashicorp/consul/api"
)

type Consul struct {
	Endpoint   string
	Datacenter string
}

type ConsulClient struct {
	client *consul.Client
	ID     string
}

func NewConsul(addr, dataCenter, serverID string) *ConsulClient {
	var (
		client *consul.Client
	)
	client, _ = consul.NewClient(&consul.Config{
		Address:    addr,
		Scheme:     "http",
		Datacenter: dataCenter,
		WaitTime:   60,
	})
	return &ConsulClient{
		client: client,
		ID:     serverID,
	}
}

func (d *ConsulClient) RegisterConsul(serviceName, host string, port, metricPort int, tags []string, metadata map[string]string) (cancelFunc context.CancelFunc) {
	var (
		id      string
		service *consul.AgentServiceRegistration
		agent   *consul.Agent
		err     error
	)
	// prometheus consul can't use other port
	if metricPort == 0 {
		metricPort = port
	}
	agent = d.client.Agent()
	id = d.ID

	if id == "" {
		id = fmt.Sprintf("%s:%d", host, port)
	}
	service = &consul.AgentServiceRegistration{
		ID:      id,
		Name:    serviceName,
		Tags:    tags,
		Port:    metricPort,
		Address: host,
		TaggedAddresses: map[string]consul.ServiceAddress{
			"metrics": {
				Address: host,
				Port:    metricPort,
			},
			"service": {
				Address: host,
				Port:    port,
			},
		},
		EnableTagOverride: false,
		Meta:              metadata,
		Weights:           nil,
		Check: &consul.AgentServiceCheck{
			Interval: "30s",
			GRPC:     fmt.Sprintf("%s:%d/%s", host, port, serviceName),
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err = agent.ServiceRegister(service); err != nil {
		log.Errorf("register comet fail, on: %s", err.Error())
		cancel()
		return
	}
	ch := make(chan struct{}, 1)
	cancelFunc = context.CancelFunc(func() {
		agent.ServiceDeregister(service.ID)
		cancel()
		<-ch
	})
	go func() {
		for {
			select {
			case <-ctx.Done():
				err = agent.ServiceDeregister(service.ID)
				log.Infof("service deregister, %#v", err)
				ch <- struct{}{}
				return
			}
		}
	}()
	return
}

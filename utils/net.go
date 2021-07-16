package utils

import "net"

func GetIP() (myIp string) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, addr := range addrs {
		if ip, ok := addr.(*net.IPNet); ok {
			if ip.IP.IsLoopback() || ip.IP.To4() == nil {
				continue
			}
			myIp = ip.IP.String()
			return
		}
	}
	return
}

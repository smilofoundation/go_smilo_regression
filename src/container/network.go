package container

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

const (
	FirstOctet        = 172
	SecondOctet       = 17
	networkNamePrefix = "testnet"
)

type NetworkManager interface {
	TryGetFreeSubnet() string
}

var defaultNetworkManager = newNetworkManager()

func newNetworkManager() *networkManager {
	return &networkManager{
		secondOctet: SecondOctet,
	}
}

type networkManager struct {
	mutex       sync.Mutex
	secondOctet int
}

func (n *networkManager) TryGetFreeSubnet() string {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.secondOctet++
	return fmt.Sprintf("%d.%d.0.0/16", FirstOctet, n.secondOctet)
}

type DockerNetwork struct {
	client  *client.Client
	id      string
	name    string
	ipv4Net *net.IPNet

	mutex   sync.Mutex
	ipIndex net.IP
}

func NewDockerNetwork() (*DockerNetwork, error) {
	c, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	network := &DockerNetwork{
		client: c,
	}

	if err := network.create(); err != nil {
		return nil, err
	}

	return network, nil
}

// create creates a user-defined docker network
func (n *DockerNetwork) create() error {
	n.name = fmt.Sprintf("%s%d", networkNamePrefix, time.Now().Unix())

	var maxTryCount = 15
	var err error
	var cResp types.NetworkCreateResponse
	var subnet string
	for i := 0; i < maxTryCount && n.id == ""; i++ {
		subnet = defaultNetworkManager.TryGetFreeSubnet()
		ipam := &network.IPAM{
			Config: []network.IPAMConfig{
				{
					Subnet: subnet,
				},
			},
		}
		cResp, err = n.client.NetworkCreate(context.Background(), n.name, types.NetworkCreate{
			IPAM: ipam,
		})
		if err == nil {
			break
		}
	}

	if err != nil {
		return err
	}
	n.id = cResp.ID
	_, n.ipv4Net, err = net.ParseCIDR(subnet)
	if err != nil {
		return err
	}
	// IP starts with xxx.xxx.0.1
	// Because xxx.xxx.0.1 is reserved for default Gateway IP
	n.ipIndex = net.IPv4(n.ipv4Net.IP[0], n.ipv4Net.IP[1], 0, 1)
	return nil
}

func (n *DockerNetwork) ID() string {
	return n.id
}

func (n *DockerNetwork) Name() string {
	return n.name
}

func (n *DockerNetwork) Subnet() string {
	return n.ipv4Net.String()
}

func (n *DockerNetwork) Remove() error {
	return n.client.NetworkRemove(context.Background(), n.id)
}

func (n *DockerNetwork) GetFreeIPAddrs(num int) ([]net.IP, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	ips := make([]net.IP, 0)
	for len(ips) < num && n.ipv4Net.Contains(n.ipIndex) {
		ip := dupIP(n.ipIndex)
		for j := len(ip) - 1; j >= 0; j-- {
			ip[j]++
			if ip[j] > 0 {
				break
			}
		}
		n.ipIndex = ip
		ips = append(ips, ip)
	}

	if len(ips) != num {
		return nil, errors.New("insufficient IP address.")
	}
	return ips, nil
}

func dupIP(ip net.IP) net.IP {
	// To save space, try and only use 4 bytes
	if x := ip.To4(); x != nil {
		ip = x
	}
	dup := make(net.IP, len(ip))
	copy(dup, ip)
	return dup
}

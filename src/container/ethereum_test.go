package container

import (
	"testing"

	"github.com/docker/docker/client"
	"github.com/phayes/freeport"
)

func TestEthereumContainer(t *testing.T) {
	t.SkipNow()

	dockerClient, err := client.NewEnvClient()
	if err != nil {
		t.Error(err)
	}

	geth := NewEthereum(
		dockerClient,
		ImageRepository("localhost:5000/go-smilo"),
		ImageTag("latest"),
		DataDir("/data"),
		WebSocket(),
		WebSocketAddress("0.0.0.0"),
		WebSocketAPI("admin,eth,net,web3,personal"),
		HostWebSocketPort(freeport.GetPort()),
		WebSocketOrigin("*"),
		NoDiscover(),
	)

	err = geth.Start()
	if err != nil {
		t.Error(err)
	}

	if !geth.Running() {
		t.Error("geth should be running")
	}

	err = geth.Stop()
	if err != nil {
		t.Error(err)
	}
}

package container

import (
	"testing"
	"time"
)

func TestEthereumBlockchain(t *testing.T) {
	t.SkipNow()

	dockerNetwork, err := NewDockerNetwork()
	if err != nil {
		t.Error(err)
	}
	defer dockerNetwork.Remove()

	chain := NewBlockchain(
		dockerNetwork,
		4,
		ImageRepository("localhost:5000/go-smilo"),
		ImageTag("latest"),
		DataDir("/data"),
		WebSocket(),
		WebSocketAddress("0.0.0.0"),
		WebSocketAPI("admin,eth,net,web3,personal"),
		WebSocketOrigin("*"),
		NoDiscover(),
		Password("password.txt"),
		Logging(false),
	)
	defer chain.Finalize()

	err = chain.Start(true)
	if err != nil {
		t.Error(err)
	}

	time.Sleep(5 * time.Second)

	err = chain.Stop(false)
	if err != nil {
		t.Error(err)
	}
}

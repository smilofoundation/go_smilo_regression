package container

import (
	"testing"
	"time"
)

func TestEthereumBlockchain(t *testing.T) {

	dockerNetwork, err := NewDockerNetwork()
	if err != nil {
		t.Error(err)
	}
	defer dockerNetwork.Remove()

	chain, err := NewBlockchain(
		dockerNetwork,
		4,
		ImageRepository(GetGoSmiloImage()),
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
	if err != nil {
		t.Error("Unable to create blockchain", err)
		t.Fail()
		return
	}
	defer chain.Finalize()

	err = chain.Start(true)
	if err != nil {
		t.Error(err)
	}

	time.Sleep(5 * time.Second)

	err = chain.Stop(true)
	if err != nil {
		t.Error(err)
	}
}

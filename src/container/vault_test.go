package container

import (
	"testing"

	"github.com/docker/docker/client"
	"github.com/phayes/freeport"
)

func TestVaultContainer(t *testing.T) {
	t.SkipNow()

	dockerClient, err := client.NewEnvClient()
	if err != nil {
		t.Error(err)
	}

	dockerNetwork, err := NewDockerNetwork()
	if err != nil {
		t.Error(err)
	}

	ips, err := dockerNetwork.GetFreeIPAddrs(1)
	if err != nil {
		t.Error(err)
	}
	ip := ips[0]

	port := freeport.GetPort()

	ct := NewVault(dockerClient,
		CTImageRepository("localhost:5000/vault"),
		CTImageTag("latest"),
		CTHost(ip, port),
		CTDockerNetworkName(dockerNetwork.Name()),
		CTWorkDir("/data"),
		CTLogging(false),
		CTKeyName("node"),
		CTSocketFilename("node.ipc"),
		CTVerbosity(3),
	)

	_, err = ct.GenerateKey()
	if err != nil {
		t.Error(err)
	}

	err = ct.Start()
	if err != nil {
		t.Error(err)
	}

	if !ct.Running() {
		t.Error("vault should be running")
	}

	err = ct.Stop()
	if err != nil {
		t.Error(err)
	}

	err = dockerNetwork.Remove()
	if err != nil {
		t.Error(err)
	}
}

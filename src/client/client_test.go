package client

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

func ExampleStartMining() {
	url := "ws://127.0.0.1:49733"
	client, err := Dial(url)
	if err != nil {
		log.Error("Failed to dial", "url", url, "err", err)
		return
	}

	err = client.StartMining(context.Background())
	if err != nil {
		log.Error("Failed to get fullnodes", "err", err)
		return
	}
}

func ExampleSpeakerFullnode() {
	url := "http://127.0.0.1:8547"
	client, err := Dial(url)
	if err != nil {
		log.Error("Failed to dial", "url", url, "err", err)
		return
	}

	err = client.ProposeFullnode(context.Background(), common.HexToAddress("0x6B58EA55d051008822Cf3acd684914c83aF2f588"), true)
	if err != nil {
		log.Error("Failed to get fullnodes", "err", err)
		return
	}
}

func ExampleGetFullnodes() {
	url := "ws://127.0.0.1:53257"
	client, err := Dial(url)
	if err != nil {
		log.Error("Failed to dial", "url", url, "err", err)
		return
	}

	addrs, err := client.GetFullnodes(context.Background(), nil)
	if err != nil {
		log.Error("Failed to get fullnodes", "err", err)
		return
	}
	for _, addr := range addrs {
		log.Info("address", "hex", addr.Hex())
	}
}

func ExampleAdminPeers() {
	url := "ws://127.0.0.1:62975"
	client, err := Dial(url)
	if err != nil {
		log.Error("Failed to dial", "url", url, "err", err)
		return
	}

	peersInfo, err := client.AdminPeers(context.Background())
	if err != nil {
		log.Error("Failed to get fullnodes", "err", err)
		return
	}

	log.Info("Peers connected", "peers", peersInfo, "len", len(peersInfo))
}

func ExampleAddPeer() {
	url := "ws://127.0.0.1:62975"
	client, err := Dial(url)
	if err != nil {
		log.Error("Failed to dial", "url", url, "err", err)
		return
	}

	err = client.AddPeer(context.Background(), "enode://ad5b4b201cc0ef5cd6ce27e32c223d1852a8b7d6069de5c3c597601e94841a5811a354261726da7b8f851e9042d5aeaed580dbb7493d22a5d922206dce3ccdb8@192.168.99.100:63040?discport=0")
	if err != nil {
		log.Error("Failed to get fullnodes", "err", err)
		return
	}
}

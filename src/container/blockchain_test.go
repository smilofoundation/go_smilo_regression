// Copyright 2020 smilofoundation/regression Authors
// Copyright 2019 smilofoundation/regression Authors
// Copyright 2017 AMIS Technologies
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

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

	chain := NewBlockchain(
		dockerNetwork,
		4,
		ImageRepository("quay.io/smilo/go-smilo"),
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

	err = chain.Stop(true)
	if err != nil {
		t.Error(err)
	}
}

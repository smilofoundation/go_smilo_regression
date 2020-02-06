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

package service

import (
	"bytes"
	"fmt"
	"text/template"
)

type Smilo struct {
	*Fullnode
	Vault *Vault
}

func NewSmilo(v *Fullnode, c *Vault) *Smilo {
	return &Smilo{
		Fullnode: v,
		Vault:    c,
	}
}
func (q Smilo) String() string {
	tmpl, err := template.New("smilo").Parse(smiloTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template, %v", err)
		return ""
	}

	result := new(bytes.Buffer)
	err = tmpl.Execute(result, q)
	if err != nil {
		fmt.Printf("Failed to render template, %v", err)
		return ""
	}

	return result.String()
}

var smiloTemplate = `{{ .Name }}:
    hostname: {{ .Name }}
    image: {{ .ImageName }}
    ports:
      - '{{ .Port }}:30303'
      - '{{ .RPCPort }}:8545'
    volumes:
      - {{ .Identity }}:{{ .Vault.Folder }}:z
    depends_on:
      - {{ .Vault.Name }}
    environment:
      - PRIVATE_CONFIG={{ .Vault.ConfigPath }}
    entrypoint:
      - /bin/sh
      - -c
      - |
        mkdir -p /eth/geth
        echo '{{ .Genesis }}' > /eth/genesis.json
        echo '{{ .StaticNodes }}' > /eth/geth/static-nodes.json
        geth --datadir "/eth" init "/eth/genesis.json"
        geth \
        --identity "{{ .Name }}" \
        --rpc \
        --rpcaddr "0.0.0.0" \
        --rpcport "8545" \
        --rpccorsdomain "*" \
        --datadir "/eth" \
        --port "30303" \
        --rpcapi "admin,db,eth,debug,miner,net,shh,txpool,personal,web3,smilobft,sport" \
        --networkid "2017" \
        --nat "any" \
        --nodekeyhex "{{ .NodeKey }}" \
        --mine \
        --debug \
        --metrics \
        --syncmode "full" \
        --ethstats "{{ .Name }}:{{ .EthStats }}" \
        --miner.gasprice 0
    networks:
      app_net:
        ipv4_address: {{ .IP }}
    restart: always
  {{ .Vault }}`

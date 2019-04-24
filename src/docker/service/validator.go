package service

import (
	"bytes"
	"fmt"
	"text/template"
)

type Fullnode struct {
	Identity    int
	Genesis     string
	NodeKey     string
	StaticNodes string
	Port        int
	RPCPort     int
	IP          string
	EthStats    string
	Name        string
}

func NewFullnode(identity int, genesis string, nodeKey string, staticNodes string, port int, rpcPort int, ethStats string, ip string) *Fullnode {
	return &Fullnode{
		Identity: identity,
		Genesis:  genesis,
		NodeKey:  nodeKey,
		Port:     port,
		RPCPort:  rpcPort,
		EthStats: ethStats,
		IP:       ip,
		Name:     fmt.Sprintf("fullnode-%v", identity),
	}
}

func (v Fullnode) String() string {
	tmpl, err := template.New("fullnode").Parse(fullnodeTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template, %v", err)
		return ""
	}

	result := new(bytes.Buffer)
	err = tmpl.Execute(result, v)
	if err != nil {
		fmt.Printf("Failed to render template, %v", err)
		return ""
	}

	return result.String()
}

var fullnodeTemplate = `{{ .Name }}:
    hostname: {{ .Name }}
    image: localhost:5000/go-smilo:latest
    ports:
      - '{{ .Port }}:30303'
      - '{{ .RPCPort }}:8545'
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
        --rpcapi "db,eth,net,web3,sport,personal" \
        --networkid "2017" \
        --nat "any" \
        --nodekeyhex "{{ .NodeKey }}" \
        --mine \
        --debug \
        --metrics \
        --syncmode "full" \
        --ethstats "{{ .Name }}:{{ .EthStats }}" \
        --gasprice 0
    networks:
      app_net:
        ipv4_address: {{ .IP }}
    restart: always`

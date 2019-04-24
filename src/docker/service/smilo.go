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
    image: localhost:5000/go-smilo:latest
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
    restart: always
  {{ .Vault }}`

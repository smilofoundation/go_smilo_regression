package service

import (
	"bytes"
	"fmt"
	"text/template"
)

type EthStats struct {
	Secret string
	IP     string
}

func NewEthStats(ip string, secret string) *EthStats {
	return &EthStats{
		IP:     ip,
		Secret: secret,
	}
}

func (c EthStats) Host() string {
	return fmt.Sprintf("%v@%v:3000", c.Secret, c.IP)
}

func (c EthStats) String() string {
	tmpl, err := template.New("eth_stats").Parse(ethStatsTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template, %v", err)
		return ""
	}

	result := new(bytes.Buffer)
	err = tmpl.Execute(result, c)
	if err != nil {
		fmt.Printf("Failed to render template, %v", err)
		return ""
	}

	return result.String()
}

var ethStatsTemplate = `eth-stats:
    image: quay.io/smilo/go-smilo:latest
    ports:
      - '3000:3000'
    environment:
      - WS_SECRET={{ .Secret }}
    restart: always
    networks:
      app_net:
        ipv4_address: {{ .IP }}`

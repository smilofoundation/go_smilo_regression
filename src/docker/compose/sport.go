package compose

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"go-smilo/src/blockchain/regression/src/docker/service"
)

type Compose interface {
	String() string
}

type sport struct {
	IPPrefix string
	EthStats *service.EthStats
	Services []*service.Fullnode
}

func New(ipPrefix string, number int, secret string, nodeKeys []string,
	genesis string, staticNodes string, smilo bool) Compose {
	ist := &sport{
		IPPrefix: ipPrefix,
		EthStats: service.NewEthStats(fmt.Sprintf("%v.9", ipPrefix), secret),
	}
	ist.init(number, nodeKeys, genesis, staticNodes)
	if smilo {
		return newSmilo(ist, number)
	}
	return ist
}

func (ist *sport) init(number int, nodeKeys []string, genesis string, staticNodes string) {
	for i := 0; i < number; i++ {
		s := service.NewFullnode(i,
			genesis,
			nodeKeys[i],
			"",
			30303+i,
			8545+i,
			ist.EthStats.Host(),
			// from subnet ip 10
			fmt.Sprintf("%v.%v", ist.IPPrefix, i+10),
		)

		staticNodes = strings.Replace(staticNodes, "0.0.0.0", s.IP, 1)
		ist.Services = append(ist.Services, s)
	}

	// update static nodes
	for i := range ist.Services {
		ist.Services[i].StaticNodes = staticNodes
	}
}

func (ist sport) String() string {
	tmpl, err := template.New("sport").Parse(sportTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template, %v", err)
		return ""
	}

	result := new(bytes.Buffer)
	err = tmpl.Execute(result, ist)
	if err != nil {
		fmt.Printf("Failed to render template, %v", err)
		return ""
	}

	return result.String()
}

var sportTemplate = `version: '3'
services:
  {{ .EthStats }}
  {{- range .Services }}
  {{ . }}
  {{- end }}
networks:
  app_net:
    driver: bridge
    ipam:
      driver: default
      config:
      - subnet: {{ .IPPrefix }}.0/24`

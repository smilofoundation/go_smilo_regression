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

package compose

import (
	"bytes"
	"fmt"
	"text/template"

	"go-smilo/src/blockchain/regression/src/docker/service"
)

type smilo struct {
	*sport
	Number        int
	SmiloServices []*service.Smilo
}

func newSmilo(ist *sport, number int) Compose {
	q := &smilo{
		sport:  ist,
		Number: number,
	}
	q.init()
	return q
}

func (q *smilo) init() {
	// set vaults
	var vaults []*service.Vault
	for i := 0; i < q.Number; i++ {
		vaults = append(vaults,
			service.NewVault(q.Services[i].Identity,
				// from subnet ip 100
				fmt.Sprintf("%v.%v", q.IPPrefix, i+100),
				10000+i,
			),
		)
	}
	for i := 0; i < q.Number; i++ {
		// set othernodes
		var nodes []string
		for j := 0; j < q.Number; j++ {
			if i != j {
				nodes = append(nodes, vaults[j].Host())
			}
		}
		vaults[i].SetOtherNodes(nodes)

		// update smilo service
		q.SmiloServices = append(q.SmiloServices,
			service.NewSmilo(q.Services[i], vaults[i]))
	}
}

func (q *smilo) String() string {
	tmpl, err := template.New("smilo").Funcs(template.FuncMap(
		map[string]interface{}{
			"PrintVolumes": func() (result string) {
				for i := 0; i < q.Number; i++ {
					result += fmt.Sprintf("  \"%v\":\n", i)
				}
				return
			},
		})).Parse(smiloTemplate)
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

var smiloTemplate = `version: '3'
services:
  {{ .EthStats }}
  {{- range .SmiloServices }}
  {{ . }}
  {{- end }}
networks:
  app_net:
    driver: bridge
    ipam:
      driver: default
      config:
      - subnet: {{ .IPPrefix }}.0/24
volumes:
{{ PrintVolumes }}
`

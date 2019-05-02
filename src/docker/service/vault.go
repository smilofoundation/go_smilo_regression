package service

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type Vault struct {
	Identity   int
	Name       string
	IP         string
	Port       int
	OtherNodes string
	PublicKey  string
	PrivateKey string
	SocketPath string
	ConfigPath string
	Folder     string
	KeyPath    string
}

func NewVault(identity int, ip string, port int) *Vault {
	folder := "/vault"
	keyPath := fmt.Sprintf("%v/tm", folder)
	return &Vault{
		Identity:   identity,
		Name:       fmt.Sprintf("vault-%v", identity),
		IP:         ip,
		Port:       port,
		PublicKey:  fmt.Sprintf("%v.pub", keyPath),
		PrivateKey: fmt.Sprintf("%v.key", keyPath),
		SocketPath: fmt.Sprintf("%v.ipc", keyPath),
		ConfigPath: fmt.Sprintf("%v.conf", keyPath),
		Folder:     folder,
		KeyPath:    keyPath,
	}
}

func (c *Vault) SetOtherNodes(nodes []string) {
	c.OtherNodes = strings.Join(nodes, ",")
}

func (c Vault) Host() string {
	return fmt.Sprintf("http://%v:%v/", c.IP, c.Port)
}

func (c Vault) String() string {
	tmpl, err := template.New("vault").Parse(vaultTemplate)
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

var vaultTemplate = `{{ .Name }}:
    hostname: {{ .Name }}
    image: quay.io/smilo/smilo-blackbox:latest
    ports:
      - '{{ .Port }}:{{ .Port }}'
    volumes:
      - {{ .Identity }}:{{ .Folder }}:z
      - .:/tmp/
    entrypoint:
      - /bin/sh
      - -c
      - |
        mkdir -p {{ .Folder }}
        echo "socket=\"{{ .SocketPath }}\"\npublickeys=[\"{{ .PublicKey }}\"]\n" > {{ .ConfigPath }}
        vault-node --generatekeys={{ .KeyPath }}
        cp {{ .KeyPath }}.pub /tmp/tm{{ .Identity }}.pub
        vault-node \
          --url={{ .Host }} \
          --port={{ .Port }} \
          --socket={{ .SocketPath }} \
          --othernodes={{ .OtherNodes }} \
          --publickeys={{ .PublicKey }} \
          --privatekeys={{ .PrivateKey }} \
          --storage={{ .Folder }} \
          --verbosity=4
    networks:
      app_net:
        ipv4_address: {{ .IP }}
    restart: always`

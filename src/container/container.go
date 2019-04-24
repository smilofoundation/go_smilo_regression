package container

import (
	"net"
	"net/url"
	"os"

	"github.com/docker/docker/client"
)

func (eth *ethereum) Image() string {
	if eth.imageTag == "" {
		return eth.imageRepository + ":latest"
	}
	return eth.imageRepository + ":" + eth.imageTag
}

func (eth *ethereum) ContainerID() string {
	return eth.containerID
}

func (eth *ethereum) Host() string {
	var host string
	daemonHost := os.Getenv("DOCKER_HOST")
	if daemonHost == "" {
		daemonHost = client.DefaultDockerHost
	}
	url, err := url.Parse(daemonHost)
	if err != nil {
		log.Error("Failed to parse daemon host", "host", daemonHost, "err", err)
		return host
	}

	if url.Scheme == "unix" {
		host = "localhost"
	} else {
		host, _, err = net.SplitHostPort(url.Host)
		if err != nil {
			log.Error("Failed to split host and port", "url", url.Host, "err", err)
			return ""
		}
	}

	return host
}

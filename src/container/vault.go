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
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"

	"go-smilo/src/blockchain/regression/src/common"
)

//TODO: refactor this with ethereum options?
/**
 * Vault options
 **/
type VaultOption func(*vault)

func CTImageRepository(repository string) VaultOption {
	return func(ct *vault) {
		ct.imageRepository = repository
	}
}

func CTImageTag(tag string) VaultOption {
	return func(ct *vault) {
		ct.imageTag = tag
	}
}

func CTHost(ip net.IP, port int) VaultOption {
	return func(ct *vault) {
		ct.port = fmt.Sprintf("%d", port)
		ct.ip = ip.String()
		ct.flags = append(ct.flags, fmt.Sprintf("--port=%d", port))
		ct.flags = append(ct.flags, fmt.Sprintf("--hostname=%s", ct.Host()))
	}
}

func CTLogging(enabled bool) VaultOption {
	return func(ct *vault) {
		ct.logging = enabled
	}
}

func CTDockerNetworkName(dockerNetworkName string) VaultOption {
	return func(ct *vault) {
		ct.dockerNetworkName = dockerNetworkName
	}
}

func CTWorkDir(workDir string) VaultOption {
	return func(ct *vault) {
		ct.workDir = workDir
		ct.flags = append(ct.flags, fmt.Sprintf("--storage=%s", workDir))
	}
}

func CTKeyName(keyName string) VaultOption {
	return func(ct *vault) {
		ct.keyName = keyName
		ct.flags = append(ct.flags, fmt.Sprintf("--privatekeys=%s", ct.keyPath("key")))
		ct.flags = append(ct.flags, fmt.Sprintf("--publickeys=%s", ct.keyPath("pub")))
	}
}

func CTSocketFilename(socketFilename string) VaultOption {
	return func(ct *vault) {
		ct.socketFilename = socketFilename
		ct.flags = append(ct.flags, fmt.Sprintf("--socket=%s", filepath.Join(ct.workDir, socketFilename)))
	}
}

func CTOtherNodes(urls []string) VaultOption {
	return func(ct *vault) {
		ct.flags = append(ct.flags, fmt.Sprintf("--othernodes=%s", strings.Join(urls, ",")))
	}
}

/**
 * Vault interface and constructors
 **/
type Vault interface {
	// GenerateKey() generates private/public key pair
	GenerateKey() (string, error)
	// Start() starts vault service
	Start() error
	// Stop() stops vault service
	Stop() error
	// Host() returns vault service url
	Host() string
	// Running() returns true if container is running
	Running() bool
	// WorkDir() returns local working directory
	WorkDir() string
	// ConfigPath() returns container config path
	ConfigPath() string
	// Binds() returns volume binding paths
	Binds() []string
	// PublicKeys() return public keys
	PublicKeys() []string
}

func NewVault(c *client.Client, options ...VaultOption) *vault {
	ct := &vault{
		client: c,
	}

	for _, opt := range options {
		opt(ct)
	}

	filters := filters.NewArgs()
	filters.Add("reference", ct.Image())

	images, err := c.ImageList(context.Background(), types.ImageListOptions{
		Filters: filters,
	})

	if len(images) == 0 || err != nil {
		out, err := ct.client.ImagePull(context.Background(), ct.Image(), types.ImagePullOptions{})
		if err != nil {
			log.Error("Failed to pull image", "image", ct.Image(), "err", err)
			return nil
		}
		if ct.logging {
			io.Copy(os.Stdout, out)
		} else {
			io.Copy(ioutil.Discard, out)
		}
	}

	return ct
}

/**
 * Vault implementation
 **/
type vault struct {
	flags          []string
	ip             string
	port           string
	containerID    string
	workDir        string
	localWorkDir   string
	keyName        string
	socketFilename string

	imageRepository   string
	imageTag          string
	dockerNetworkName string

	logging bool
	client  *client.Client
}

func (ct *vault) Image() string {
	if ct.imageTag == "" {
		return ct.imageRepository + ":latest"
	}
	return ct.imageRepository + ":" + ct.imageTag
}

func (ct *vault) GenerateKey() (localWorkDir string, err error) {
	// Generate empty password file
	ct.localWorkDir, err = common.GenerateRandomDir()
	if err != nil {
		log.Error("Failed to generate working dir", "dir", ct.localWorkDir, "err", err)
		return "", err
	}

	// Generate config file
	configContent := fmt.Sprintf("socket=\"%s\"\npublickeys=[\"%s\"]\n",
		ct.keyPath("ipc"), ct.keyPath("pub"))
	localConfigPath := ct.localConfigPath()
	err = ioutil.WriteFile(localConfigPath, []byte(configContent), 0600)
	if err != nil {
		log.Error("Failed to write config", "file", localConfigPath, "err", err)
		return "", err
	}

	// Create container and mount working directory
	binds := ct.Binds()
	config := &container.Config{
		Image: ct.Image(),
		Cmd: []string{
			"--generate-keys=" + ct.keyPath(""),
		},
	}
	hostConfig := &container.HostConfig{
		Binds: binds,
	}
	resp, err := ct.client.ContainerCreate(context.Background(), config, hostConfig, nil, "")
	if err != nil {
		log.Error("Failed to create container", "err", err)
		return "", err
	}
	id := resp.ID

	// Start container
	if err := ct.client.ContainerStart(context.Background(), id, types.ContainerStartOptions{}); err != nil {
		log.Error("Failed to start container", "err", err)
		return "", err
	}

	// Attach container: for stdin interaction with the container.
	// - vault-node generatekeys takes stdin as password
	hiresp, err := ct.client.ContainerAttach(context.Background(), id, types.ContainerAttachOptions{Stream: true, Stdin: true})
	if err != nil {
		log.Error("Failed to attach container", "err", err)
		return "", err
	}
	// - write empty string password to container stdin
	hiresp.Conn.Write([]byte("")) //Empty password

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	waitC, _ := ct.client.ContainerWait(ctx, id, container.WaitConditionNotRunning)

	if status := <-waitC; status.StatusCode != 0 {
		//close(finished)
		err1 := fmt.Errorf("a non-zero code from VAULT ContainerWait: %d", status.StatusCode)
		logCancellationError(err1.Error())
		return "", err1
	}
	log.Info("Managed to start VAULT container ", "id", id)

	ctx1, cancel1 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel1()

	err = ct.client.ContainerKill(ctx1, id, "")
	if err != nil {
		log.Error("Failed to kill VAULT container", "err", err)
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel2()

	err = ct.client.ContainerRemove(ctx2, id, types.ContainerRemoveOptions{Force: true})
	if err != nil {
		log.Error("Failed to remove GETH container", "err", err)
	}

	return "", nil
}

func (ct *vault) Start() error {
	defer func() {
		if ct.logging {
			go ct.showLog(context.Background())
		}
	}()

	// container config
	exposedPorts := make(map[nat.Port]struct{})
	exposedPorts[nat.Port(ct.port)] = struct{}{}
	config := &container.Config{
		Image:        ct.Image(),
		Cmd:          ct.flags,
		ExposedPorts: exposedPorts,
	}

	// host config
	binds := []string{
		ct.localWorkDir + ":" + ct.workDir,
	}
	hostConfig := &container.HostConfig{
		Binds: binds,
	}

	// Setup network config
	var networkingConfig *network.NetworkingConfig
	if ct.ip != "" && ct.dockerNetworkName != "" {
		endpointsConfig := make(map[string]*network.EndpointSettings)
		endpointsConfig[ct.dockerNetworkName] = &network.EndpointSettings{
			IPAMConfig: &network.EndpointIPAMConfig{
				IPv4Address: ct.ip,
			},
		}
		networkingConfig = &network.NetworkingConfig{
			EndpointsConfig: endpointsConfig,
		}
	}

	// Create container
	resp, err := ct.client.ContainerCreate(context.Background(), config, hostConfig, networkingConfig, "")
	if err != nil {
		log.Error("Failed to create container", "err", err)
		return err
	}
	ct.containerID = resp.ID

	// Start container
	err = ct.client.ContainerStart(context.Background(), ct.containerID, types.ContainerStartOptions{})
	if err != nil {
		log.Error("Failed to start container", "ip", ct.ip, "err", err)
		return err
	}

	return nil
}

func (ct *vault) Stop() error {
	err := ct.client.ContainerStop(context.Background(), ct.containerID, nil)
	if err != nil {
		return err
	}

	defer os.RemoveAll(ct.localWorkDir)

	return ct.client.ContainerRemove(context.Background(), ct.containerID,
		types.ContainerRemoveOptions{
			Force: true,
		})
}

func (ct *vault) Host() string {
	return fmt.Sprintf("http://%s:%s/", ct.ip, ct.port)
}

func (ct *vault) Running() bool {
	containers, err := ct.client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Error("Failed to list containers", "err", err)
		return false
	}

	for _, c := range containers {
		if c.ID == ct.containerID {
			return true
		}
	}

	return false
}

func (ct *vault) WorkDir() string {
	return ct.localWorkDir
}

func (ct *vault) ConfigPath() string {
	return ct.keyPath("conf")
}

func (ct *vault) Binds() []string {
	return []string{ct.localWorkDir + ":" + ct.workDir}
}

func (ct *vault) PublicKeys() []string {
	keyPath := ct.localKeyPath("pub")
	keyBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Error("Unable to read key file", "file", keyPath, "err", err)
		return nil
	}
	return []string{string(keyBytes)}
}

/**
 * Vault internal functions
 **/

func (ct *vault) showLog(context context.Context) {
	if readCloser, err := ct.client.ContainerLogs(context, ct.containerID,
		types.ContainerLogsOptions{ShowStderr: true, Follow: true}); err == nil {
		defer readCloser.Close()
		_, err = io.Copy(os.Stdout, readCloser)
		if err != nil && err != io.EOF {
			log.Error("Failed to print container log", "err", err)
			return
		}
	}
}

func (ct *vault) keyPath(extension string) string {
	if extension == "" {
		return filepath.Join(ct.workDir, ct.keyName)
	} else {
		return filepath.Join(ct.workDir, fmt.Sprintf("%s.%s", ct.keyName, extension))
	}
}

func (ct *vault) localKeyPath(extension string) string {
	return filepath.Join(ct.localWorkDir, fmt.Sprintf("%s.%s", ct.keyName, extension))
}

func (ct *vault) localConfigPath() string {
	return filepath.Join(ct.localWorkDir, fmt.Sprintf("%s.conf", ct.keyName))
}

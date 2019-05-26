package container

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/docker/docker/client"
	"github.com/ethereum/go-ethereum/common"
	"github.com/phayes/freeport"

	"go-smilo/src/blockchain/smilobft/accounts"
	"go-smilo/src/blockchain/smilobft/accounts/keystore"

	smilocommon "go-smilo/src/blockchain/regression/src/common"
	"go-smilo/src/blockchain/regression/src/genesis"
)

const (
	allocBalance     = "900000000000000000000000000000000000000000000"
	veryLightScryptN = 2
	veryLightScryptP = 1
	defaultPassword  = ""
)

type NodeIncubator interface {
	CreateNodes(int, ...Option) ([]Ethereum, error)
}

type Blockchain interface {
	AddFullnodes(numOfFullnodes int) ([]Ethereum, error)
	RemoveFullnodes(candidates []Ethereum, t time.Duration) error
	EnsureConsensusWorking(geths []Ethereum, t time.Duration) error
	Start(bool) error
	Stop(bool) error
	Fullnodes() []Ethereum
	Finalize()
}

func GetGoSmiloImage() string {
	return "quay.io/smilo/go-smilo"
}

func GetVaultImage() string {
	return "quay.io/smilo/smilo-blackbox"
}
func NewBlockchain(network *DockerNetwork, numOfFullnodes int, options ...Option) (bc *blockchain, err error) {
	if network == nil {
		//log.Error("Docker network is required")
		return nil, fmt.Errorf("Docker network is required")
	}

	bc = &blockchain{dockerNetwork: network, opts: options}

	var err1 error
	bc.dockerClient, err1 = client.NewEnvClient()
	if err1 != nil {
		//log.Error("Failed to connect to Docker daemon", "err", err)
		return nil, fmt.Errorf("Failed to connect to Docker daemon %s", err1)
	}

	bc.opts = append(bc.opts, DockerNetworkName(bc.dockerNetwork.Name()))

	//Create accounts
	bc.generateAccounts(numOfFullnodes)

	if err1 = bc.addFullnodes(numOfFullnodes); err1 != nil {
		//log.Error("Error creating fullnodes", "err", err1)
		return nil, fmt.Errorf("Error creating fullnodes %s", err1)
	}
	return bc, nil
}

func NewDefaultBlockchain(network *DockerNetwork, numOfFullnodes int) (bc *blockchain, err error) {
	return NewBlockchain(network,
		numOfFullnodes,
		ImageRepository(GetGoSmiloImage()),
		ImageTag("latest"),
		DataDir("/data"),
		WebSocket(),
		WebSocketAddress("0.0.0.0"),
		WebSocketAPI("personal,admin,db,eth,net,web3,miner,shh,txpool,debug,smilobft,sport"),
		WebSocketOrigin("*"),
		RPC(),
		RPCAddress("0.0.0.0"),
		RPCAPI("personal,admin,db,eth,debug,miner,net,shh,txpool,personal,web3,smilobft,sport"),
		RPCOrigin("*"),
		NAT("any"),
		NoDiscover(),
		//Testnet(),
		Etherbase("1a9afb711302c5f83b5902843d1c007a1a137632"),
		Mine(),
		SyncMode("full"),
		Unlock(0),
		Password("password.txt"),
		Logging(true),
		IsSmilo(true),
	)
}

func NewDefaultBlockchainWithFaulty(network *DockerNetwork, numOfNormal int, numOfFaulty int) (bc *blockchain, err error) {
	if network == nil {
		//log.Error("Docker network is required")
		return nil, fmt.Errorf("Docker network is required")
	}

	commonOpts := [...]Option{
		DockerNetworkName(network.Name()),
		DataDir("/data"),
		WebSocket(),
		WebSocketAddress("0.0.0.0"),
		WebSocketAPI("personal,admin,db,eth,net,web3,miner,shh,txpool,debug,smilobft,sport"),
		WebSocketOrigin("*"),
		RPC(),
		RPCAddress("0.0.0.0"),
		RPCAPI("personal,admin,db,eth,debug,miner,net,shh,txpool,personal,web3,smilobft,sport"),
		RPCOrigin("*"),
		NAT("any"),
		NoDiscover(),
		//Testnet(),
		Etherbase("1a9afb711302c5f83b5902843d1c007a1a137632"),
		Mine(),
		SyncMode("full"),
		Unlock(0),
		Password("password.txt"),
		Logging(false)}
	normalOpts := make([]Option, len(commonOpts), len(commonOpts)+2)
	copy(normalOpts, commonOpts[:])
	normalOpts = append(normalOpts, ImageRepository(GetGoSmiloImage()), ImageTag("latest"))
	faultyOpts := make([]Option, len(commonOpts), len(commonOpts)+3)
	copy(faultyOpts, commonOpts[:])
	faultyOpts = append(faultyOpts, ImageRepository(GetGoSmiloImage()), ImageTag("regression_test"), FaultyMode(1))

	// New env client
	bc = &blockchain{dockerNetwork: network}
	var err1 error
	bc.dockerClient, err1 = client.NewEnvClient()
	if err1 != nil {
		//log.Error("Failed to connect to Docker daemon", "err", err)
		return nil, fmt.Errorf("Failed to connect to Docker daemon %s", err1)
	}

	totalNodes := numOfNormal + numOfFaulty

	ips, err1 := bc.dockerNetwork.GetFreeIPAddrs(totalNodes)
	if err1 != nil {
		//log.Error("Failed to get free ip addresses", "err", err)
		return nil, fmt.Errorf("Failed to get free ip addresses %s", err1)
	}

	//Create accounts
	bc.generateAccounts(totalNodes)

	keys, _, addrs := smilocommon.GenerateKeys(totalNodes)
	bc.setupGenesis(addrs)
	// Create normal fullnodes
	bc.opts = normalOpts
	if err1 = bc.setupFullnodes(ips[:numOfNormal], keys[:numOfNormal], 0, bc.opts...); err1 != nil {
		//log.Error("Error setting up normal fullnodes")
		return nil, fmt.Errorf("Error setting up normal fullnodes %s", err1)
	}
	// Create faulty fullnodes
	bc.opts = faultyOpts
	if err1 = bc.setupFullnodes(ips[numOfNormal:], keys[numOfNormal:], numOfNormal, bc.opts...); err1 != nil {
		//log.Error("Error setting up faulty fullnodes")
		return nil, fmt.Errorf("Error setting up faulty fullnodes %s", err1)
	}
	return bc, nil
}

func NewSmiloBlockchain(network *DockerNetwork, ctn VaultNetwork, options ...Option) (bc *blockchain, err error) {
	if network == nil {
		//log.Error("Docker network is required")
		return nil, fmt.Errorf("Docker network is required")
	}

	bc = &blockchain{dockerNetwork: network, opts: options, isSmilo: true, vaultNetwork: ctn}
	bc.opts = append(bc.opts, IsSmilo(true))
	bc.opts = append(bc.opts, NoUSB())

	var err1 error
	bc.dockerClient, err1 = client.NewEnvClient()
	if err1 != nil {
		//log.Error("Failed to connect to Docker daemon", "err", err1)
		return nil, fmt.Errorf("Failed to connect to Docker daemon %s", err1)
	}

	bc.opts = append(bc.opts, DockerNetworkName(bc.dockerNetwork.Name()))

	//Create accounts
	bc.generateAccounts(ctn.NumOfVaults())

	if err1 := bc.addFullnodes(ctn.NumOfVaults()); err1 != nil {
		//log.Error("Error creating fullnodes", "err", err1)
		return nil, fmt.Errorf("Error creating fullnodes %s", err1)
	}
	return bc, nil
}

func NewDefaultSmiloBlockchain(network *DockerNetwork, ctn VaultNetwork) (bc *blockchain, err error) {
	return NewSmiloBlockchain(network,
		ctn,
		ImageRepository(GetGoSmiloImage()),
		ImageTag("latest"),
		DataDir("/data"),
		WebSocket(),
		WebSocketAddress("0.0.0.0"),
		WebSocketAPI("personal,admin,db,eth,net,web3,miner,shh,txpool,debug,smilobft,sport"),
		WebSocketOrigin("*"),
		RPC(),
		RPCAddress("0.0.0.0"),
		RPCAPI("personal,admin,db,eth,debug,miner,net,shh,txpool,personal,web3,smilobft,sport"),
		RPCOrigin("*"),
		NAT("any"),
		NoDiscover(),
		//Testnet(),
		Etherbase("1a9afb711302c5f83b5902843d1c007a1a137632"),
		Mine(),
		SyncMode("full"),
		Unlock(0),
		Password("password.txt"),
		Logging(false),
		IsSmilo(true),
	)
}

func NewDefaultSmiloBlockchainWithFaulty(network *DockerNetwork, ctn VaultNetwork, numOfNormal int, numOfFaulty int) (bc *blockchain, err error) {
	if network == nil {
		//log.Error("Docker network is required")
		return nil, fmt.Errorf("Docker network is required")
	}

	commonOpts := [...]Option{
		DockerNetworkName(network.Name()),
		DataDir("/data"),
		WebSocket(),
		WebSocketAddress("0.0.0.0"),
		WebSocketAPI("personal,admin,db,eth,net,web3,miner,shh,txpool,debug,smilobft,sport"),
		WebSocketOrigin("*"),
		RPC(),
		RPCAddress("0.0.0.0"),
		RPCAPI("personal,admin,db,eth,debug,miner,net,shh,txpool,personal,web3,smilobft,sport"),
		RPCOrigin("*"),
		NAT("any"),
		NoDiscover(),
		//Testnet(),
		Etherbase("1a9afb711302c5f83b5902843d1c007a1a137632"),
		Mine(),
		SyncMode("full"),
		Unlock(0),
		Password("password.txt"),
		Logging(false),
		IsSmilo(true),
	}
	normalOpts := make([]Option, len(commonOpts), len(commonOpts)+2)
	copy(normalOpts, commonOpts[:])
	normalOpts = append(normalOpts, ImageRepository(GetGoSmiloImage()), ImageTag("latest"))
	faultyOpts := make([]Option, len(commonOpts), len(commonOpts)+3)
	copy(faultyOpts, commonOpts[:])
	faultyOpts = append(faultyOpts, ImageRepository(GetGoSmiloImage()), ImageTag("regression_test"), FaultyMode(1))

	// New env client
	bc = &blockchain{dockerNetwork: network, isSmilo: true, vaultNetwork: ctn}
	var err1 error
	bc.dockerClient, err1 = client.NewEnvClient()
	if err1 != nil {
		//log.Error("Failed to connect to Docker daemon", "err", err1)
		return nil, fmt.Errorf("Failed to connect to Docker daemon %s", err1)
	}

	totalNodes := numOfNormal + numOfFaulty

	ips, err1 := bc.dockerNetwork.GetFreeIPAddrs(totalNodes)
	if err1 != nil {
		//log.Error("Failed to get free ip addresses", "err", err1)
		return nil, fmt.Errorf("Failed to get free ip addresses %s", err1)
	}

	//Create accounts
	bc.generateAccounts(totalNodes)

	keys, _, addrs := smilocommon.GenerateKeys(totalNodes)
	bc.setupGenesis(addrs)
	// Create normal fullnodes
	bc.opts = normalOpts
	if err1 = bc.setupFullnodes(ips[:numOfNormal], keys[:numOfNormal], 0, bc.opts...); err1 != nil {
		//log.Error("Error setting up normal fullnodes")
		return nil, fmt.Errorf("Error setting up normal fullnodes %s", err1)
	}
	// Create faulty fullnodes
	bc.opts = faultyOpts
	if err1 = bc.setupFullnodes(ips[numOfNormal:], keys[numOfNormal:], numOfNormal, bc.opts...); err1 != nil {
		//log.Error("Error setting up faulty fullnodes")
		return nil, fmt.Errorf("Error setting up faulty fullnodes %s", err1)
	}
	return bc, nil
}

// ----------------------------------------------------------------------------

type blockchain struct {
	dockerClient  *client.Client
	dockerNetwork *DockerNetwork
	genesisFile   string
	isSmilo       bool
	fullnodes     []Ethereum
	opts          []Option
	vaultNetwork  VaultNetwork
	accounts      []accounts.Account
	keystorePath  string
}

func (bc *blockchain) AddFullnodes(numOfFullnodes int) ([]Ethereum, error) {
	// TODO: need a lock
	lastLen := len(bc.fullnodes)
	bc.addFullnodes(numOfFullnodes)

	newFullnodes := bc.fullnodes[lastLen:]
	if err := bc.start(newFullnodes); err != nil {
		return nil, err
	}

	// propose new fullnodes as fullnode in consensus
	for _, v := range bc.fullnodes[:lastLen] {
		istClient := v.NewClient()
		for _, newV := range newFullnodes {
			if err := istClient.ProposeFullnode(context.Background(), newV.Address(), true); err != nil {
				return nil, err
			}
		}
	}

	if err := bc.connectAll(true); err != nil {
		return nil, err
	}
	return newFullnodes, nil
}

func (bc *blockchain) EnsureConsensusWorking(geths []Ethereum, t time.Duration) error {
	errCh := make(chan error, len(geths))
	quitCh := make(chan struct{}, len(geths))
	for _, geth := range geths {
		go geth.ConsensusMonitor(errCh, quitCh)
	}

	timeout := time.NewTimer(t)
	defer timeout.Stop()

	var err error
	select {
	case err = <-errCh:
	case <-timeout.C:
		for i := 0; i < len(geths); i++ {
			quitCh <- struct{}{}
		}
	}
	return err
}

func (bc *blockchain) RemoveFullnodes(candidates []Ethereum, processingTime time.Duration) error {
	var newFullnodes []Ethereum

	for _, v := range bc.fullnodes {
		istClient := v.NewClient()
		isFound := false
		for _, c := range candidates {
			if err := istClient.ProposeFullnode(context.Background(), c.Address(), false); err != nil {
				return err
			}
			if v.ContainerID() == c.ContainerID() {
				isFound = true
			}
		}
		if !isFound {
			newFullnodes = append(newFullnodes, v)
		}
	}

	// FIXME: It is not good way to wait fullnode vote out candidates
	<-time.After(processingTime)
	bc.fullnodes = newFullnodes

	return bc.stop(candidates, false)
}

func (bc *blockchain) Start(strong bool) error {
	if err := bc.start(bc.fullnodes); err != nil {
		return err
	}
	return bc.connectAll(strong)
}

func (bc *blockchain) Stop(force bool) error {
	if err := bc.stop(bc.fullnodes, force); err != nil {
		return err
	}

	return nil
}

func (bc *blockchain) Finalize() {
	os.RemoveAll(filepath.Dir(bc.genesisFile))
}

func (bc *blockchain) Fullnodes() []Ethereum {
	return bc.fullnodes
}

func (bc *blockchain) CreateNodes(num int, options ...Option) (nodes []Ethereum, err error) {
	ips, err := bc.dockerNetwork.GetFreeIPAddrs(num)
	if err != nil {
		return nil, err
	}

	for i := 0; i < num; i++ {
		var opts []Option
		opts = append(opts, options...)

		// Host data directory
		dataDir, err := smilocommon.GenerateRandomDir()
		if err != nil {
			log.Error("Failed to create data dir", "dir", dataDir, "err", err)
			return nil, err
		}
		opts = append(opts, HostDataDir(dataDir))
		opts = append(opts, HostWebSocketPort(freeport.GetPort()))
		opts = append(opts, HostIP(ips[i]))
		opts = append(opts, DockerNetworkName(bc.dockerNetwork.Name()))

		geth := NewEthereum(
			bc.dockerClient,
			opts...,
		)

		err = geth.Init(bc.genesisFile)
		if err != nil {
			log.Error("Failed to init genesis", "file", bc.genesisFile, "err", err)
			return nil, err
		}

		nodes = append(nodes, geth)
	}

	return nodes, nil
}

// ----------------------------------------------------------------------------

func (bc *blockchain) addFullnodes(numOfFullnodes int) error {
	ips, err := bc.dockerNetwork.GetFreeIPAddrs(numOfFullnodes)
	if err != nil {
		return err
	}
	keys, _, addrs := smilocommon.GenerateKeys(numOfFullnodes)
	bc.setupGenesis(addrs)
	if err = bc.setupFullnodes(ips, keys, 0, bc.opts...); err != nil {
		return err
	}
	return nil
}

func (bc *blockchain) connectAll(strong bool) error {
	for idx, v := range bc.fullnodes {
		if strong {
			for _, vv := range bc.fullnodes {
				if v.ContainerID() != vv.ContainerID() {
					if err := v.AddPeer(vv.NodeAddress()); err != nil {
						return err
					}
				}
			}
		} else {
			nextFullnode := bc.fullnodes[(idx+1)%len(bc.fullnodes)]
			if err := v.AddPeer(nextFullnode.NodeAddress()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (bc *blockchain) generateAccounts(num int) {
	// Create keystore object
	d, err := ioutil.TempDir("", "sport-keystore")
	if err != nil {
		log.Error("Failed to create temp folder for keystore", "err", err)
		return
	}
	ks := keystore.NewKeyStore(d, veryLightScryptN, veryLightScryptP)
	bc.keystorePath = d

	// Create accounts
	for i := 0; i < num; i++ {
		a, e := ks.NewAccount(defaultPassword)
		if e != nil {
			log.Error("Failed to create account", "err", err)
			return
		}
		bc.accounts = append(bc.accounts, a)
	}
}

func (bc *blockchain) setupGenesis(addrs []common.Address) {
	balance, _ := new(big.Int).SetString(allocBalance, 10)
	if bc.genesisFile == "" {
		var allocAddrs []common.Address
		allocAddrs = append(allocAddrs, addrs...)
		for _, acc := range bc.accounts {
			allocAddrs = append(allocAddrs, acc.Address)
		}
		bc.genesisFile = genesis.NewFile(bc.isSmilo,
			genesis.Fullnodes(addrs...),
			genesis.Alloc(allocAddrs, balance),
		)
	}
}

// Offset: offset is for account index offset
func (bc *blockchain) setupFullnodes(ips []net.IP, keys []*ecdsa.PrivateKey, offset int, options ...Option) (error) {
	for i := 0; i < len(keys); i++ {
		var opts []Option
		opts = append(opts, options...)

		// Host data directory
		dataDir, err := smilocommon.GenerateRandomDir()
		if err != nil {
			log.Error("Failed to create data dir", "dir", dataDir, "err", err)
			return err
		}
		opts = append(opts, HostDataDir(dataDir))
		opts = append(opts, HostWebSocketPort(freeport.GetPort()))
		opts = append(opts, Key(keys[i]))
		opts = append(opts, HostIP(ips[i]))

		accounts := bc.accounts[i+offset : i+offset+1]
		var addrs []common.Address
		for _, acc := range accounts {
			addrs = append(addrs, acc.Address)
		}
		opts = append(opts, Accounts(addrs))

		// Add PRIVATE_CONFIG for smilo
		if bc.isSmilo {
			ct := bc.vaultNetwork.GetVault(i)
			env := fmt.Sprintf("PRIVATE_CONFIG=%s", ct.ConfigPath())
			opts = append(opts, DockerEnv([]string{env}))
			opts = append(opts, DockerBinds(ct.Binds()))
		}

		geth := NewEthereum(
			bc.dockerClient,
			opts...,
		)

		// Copy keystore to datadir
		smilocommon.GeneratePasswordFile(dataDir, geth.password, defaultPassword)
		smilocommon.CopyKeystore(dataDir, accounts)

		err = geth.Init(bc.genesisFile)
		if err != nil {
			log.Error("Failed to init genesis", "file", bc.genesisFile, "err", err)
			return err
		}

		bc.fullnodes = append(bc.fullnodes, geth)
	}
	return nil
}

func (bc *blockchain) start(fullnodes []Ethereum) error {
	for _, v := range fullnodes {
		if err := v.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (bc *blockchain) stop(fullnodes []Ethereum, force bool) error {
	for _, v := range fullnodes {
		if err := v.Stop(); err != nil && !force {
			return err
		}
	}
	return nil
}

// Vault functions ----------------------------------------------------------------------------
type VaultNetwork interface {
	Start() error
	Stop() error
	Finalize()
	NumOfVaults() int
	GetVault(int) Vault
}

func NewVaultNetwork(network *DockerNetwork, numOfFullnodes int, options ...VaultOption) (ctn *vaultNetwork, err error) {
	if network == nil {
		//log.Error("Docker network is required")
		return nil, fmt.Errorf("Docker network is required")
	}
	ctn = &vaultNetwork{dockerNetwork: network, opts: options}

	var err1 error
	ctn.dockerClient, err1 = client.NewEnvClient()
	if err1 != nil {
		log.Error("Failed to connect to Docker daemon", "err", err1)
		return nil, fmt.Errorf("Failed to connect to Docker daemon %s", err1)
	}

	ctn.opts = append(ctn.opts, CTDockerNetworkName(ctn.dockerNetwork.Name()))

	if err1 := ctn.setupVaults(numOfFullnodes); err1 != nil {
		return nil, fmt.Errorf("Failed to setup vaults %s", err1)
	}
	return ctn, nil
}

func NewDefaultVaultNetwork(network *DockerNetwork, numOfFullnodes int) (ctn *vaultNetwork, err error) {

	return NewVaultNetwork(network, numOfFullnodes,
		CTImageRepository(GetVaultImage()),
		CTImageTag("latest"),
		CTWorkDir("/ctdata"),
		CTLogging(false),
		CTKeyName("node"),
		CTSocketFilename("node.ipc"),
		//CTVerbosity(1),
	)
}

func (ctn *vaultNetwork) setupVaults(numOfFullnodes int) (error) {
	// Create vaultsF
	ips, ports := ctn.getFreeHosts(numOfFullnodes)
	for i := 0; i < numOfFullnodes; i++ {
		opts := append(ctn.opts, CTHost(ips[i], ports[i]))
		othernodes := ctn.getOtherNodes(ips, ports, i)
		opts = append(opts, CTOtherNodes(othernodes))
		ct := NewVault(ctn.dockerClient, opts...)
		// Generate keys
		if _, err := ct.GenerateKey(); err != nil {
			return err
		}
		ctn.vaults = append(ctn.vaults, ct)
	}
	return nil
}

func (ctn *vaultNetwork) Start() error {
	// Run nodes
	for i, ct := range ctn.vaults {
		err := ct.Start()
		if err != nil {
			log.Error("Failed to start vault", "index", i, "err", err)
			return err
		}
	}
	return nil
}

func (ctn *vaultNetwork) Stop() error {
	// Stop nodes
	for i, ct := range ctn.vaults {
		err := ct.Stop()
		if err != nil {
			log.Error("Failed to stop vault", "index", i, "err", err)
			return err
		}
	}
	return nil
}

func (ctn *vaultNetwork) Finalize() {
	// Clean up local working directory
	for _, ct := range ctn.vaults {
		os.RemoveAll(ct.WorkDir())
	}
}

func (ctn *vaultNetwork) NumOfVaults() int {
	return len(ctn.vaults)
}

func (ctn *vaultNetwork) GetVault(idx int) Vault {
	return ctn.vaults[idx]
}

func (ctn *vaultNetwork) getFreeHosts(num int) ([]net.IP, []int) {
	ips, err := ctn.dockerNetwork.GetFreeIPAddrs(num)
	if err != nil {
		log.Error("Cannot get free ip", "err", err)
		return nil, nil
	}
	var ports []int
	for i := 0; i < num; i++ {
		ports = append(ports, freeport.GetPort())
	}
	return ips, ports
}

func (ctn *vaultNetwork) getOtherNodes(ips []net.IP, ports []int, idx int) []string {
	var result []string
	for i, ip := range ips {
		if i == idx {
			continue
		}
		result = append(result, fmt.Sprintf("http://%s:%d/", ip, ports[i]))
	}
	return result
}

type vaultNetwork struct {
	dockerClient  *client.Client
	dockerNetwork *DockerNetwork
	opts          []VaultOption
	vaults        []Vault
}

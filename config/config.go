package config

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/hyperledger-labs/yui-relayer/core"
)

type Config struct {
	Global GlobalConfig             `yaml:"global" json:"global"`
	Chains []core.ChainClientConfig `yaml:"chains" json:"chains"`
	Paths  core.Paths               `yaml:"paths" json:"paths"`

	// cache
	chains Chains `yaml:"-" json:"-"`
}

func DefaultConfig() Config {
	return Config{
		Global: newDefaultGlobalConfig(),
		Chains: []core.ChainClientConfig{},
		Paths:  core.Paths{},
	}
}

// GlobalConfig describes any global relayer settings
type GlobalConfig struct {
	Timeout        string `yaml:"timeout" json:"timeout"`
	LightCacheSize int    `yaml:"light-cache-size" json:"light-cache-size"`
}

// newDefaultGlobalConfig returns a global config with defaults set
func newDefaultGlobalConfig() GlobalConfig {
	return GlobalConfig{
		Timeout:        "10s",
		LightCacheSize: 20,
	}
}

func (c *Config) GetChain(chainID string) (*core.Chain, error) {
	return c.chains.Get(chainID)
}

func (c *Config) GetChains(chainIDs ...string) (map[string]*core.Chain, error) {
	return c.chains.Gets(chainIDs...)
}

// AddChain adds an additional chain to the config
func (c *Config) AddChain(m codec.JSONCodec, config core.ChainClientConfig) error {
	chain, err := config.BuildChain()
	if err != nil {
		return err
	}
	if _, err := c.GetChain(chain.ChainID()); err == nil {
		return fmt.Errorf("chain with ID %s already exists in config", chain.ChainID())
	}
	c.Chains = append(c.Chains, config)
	c.chains = append(c.chains, chain)
	return nil
}

// AddPath adds an additional path to the config
func (c *Config) AddPath(name string, path *core.Path) (err error) {
	return c.Paths.Add(name, path)
}

// DeleteChain removes a chain from the config
func (c *Config) DeleteChain(chain string) *Config {
	var newChains []*core.Chain
	var newChainConfigs []core.ChainClientConfig
	for i, ch := range c.chains {
		if ch.ChainID() != chain {
			newChains = append(newChains, ch)
			newChainConfigs = append(newChainConfigs, c.Chains[i])
		}
	}
	c.Chains = newChainConfigs
	c.chains = newChains
	return c
}

// ChainsFromPath takes the path name and returns the properly configured chains
func (c *Config) ChainsFromPath(path string) (map[string]*core.Chain, string, string, error) {
	pth, err := c.Paths.Get(path)
	if err != nil {
		return nil, "", "", err
	}

	src, dst := pth.Src.ChainID, pth.Dst.ChainID
	chains, err := c.chains.Gets(src, dst)
	if err != nil {
		return nil, "", "", err
	}

	if err = chains[src].SetPath(pth.Src); err != nil {
		return nil, "", "", err
	}
	if err = chains[dst].SetPath(pth.Dst); err != nil {
		return nil, "", "", err
	}

	return chains, src, dst, nil
}

// Called to initialize the relayer.Chain types on Config
func InitChains(c *Config, homePath string, debug bool) error {
	to, err := time.ParseDuration(c.Global.Timeout)
	if err != nil {
		return fmt.Errorf("did you remember to run 'rly config init' error:%w", err)
	}

	for _, chain := range c.chains {
		if err := chain.Init(homePath, to, debug); err != nil {
			return fmt.Errorf("did you remember to run 'rly config init' error:%w", err)
		}
	}

	return nil
}

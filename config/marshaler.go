package config

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
)

func MarshalJSON(config Config) ([]byte, error) {
	return json.Marshal(config)
}

func UnmarshalJSON(m codec.Codec, bz []byte, config *Config) error {
	if err := json.Unmarshal(bz, config); err != nil {
		return err
	}
	for _, c := range config.Chains {
		if err := c.Init(m); err != nil {
			return err
		}
		chain, err := c.BuildChain()
		if err != nil {
			return err
		}
		config.chains = append(config.chains, chain)
	}
	return nil
}

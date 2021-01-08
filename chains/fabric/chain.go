package fabric

import (
	"fmt"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/datachainlab/fabric-ibc/x/auth/types"
	"github.com/datachainlab/relayer/core"
	"github.com/tendermint/tendermint/libs/log"
)

type Chain struct {
	config ChainConfig

	pathEnd  *core.PathEnd
	homePath string

	encodingConfig params.EncodingConfig
	gateway        FabricGateway
	logger         log.Logger
}

func NewChain(config ChainConfig) *Chain {
	return &Chain{config: config}
}

var _ core.ChainI = (*Chain)(nil)

func (c *Chain) Init(homePath string, timeout time.Duration, debug bool) error {
	c.homePath = homePath
	c.logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	c.encodingConfig = makeEncodingConfig()
	return nil
}

func (c *Chain) ClientType() string {
	return "fabric"
}

func (c *Chain) ChainID() string {
	return c.config.ChainId
}

func (c *Chain) ClientID() string {
	return c.pathEnd.ClientID
}

// GetAddress returns the sdk.AccAddress associated with the configred key
func (c *Chain) GetAddress() (sdk.AccAddress, error) {
	sid, err := c.getSerializedIdentity(c.config.MspId)
	if err != nil {
		return nil, err
	}
	return authtypes.MakeCreatorAddressWithSerializedIdentity(sid)
}

func (c *Chain) SetPath(p *core.PathEnd) error {
	err := p.Validate()
	if err != nil {
		return c.errCantSetPath(err)
	}
	c.pathEnd = p
	return nil
}

func (c *Chain) Update(key, value string) (core.ChainConfigI, error) {
	panic("not implemented error")
	return &c.config, nil
}

func (c *Chain) StartEventListener(dst core.ChainI, strategy core.StrategyI) {
	panic("not implemented error")
}

func (c *Chain) QueryLatestHeader() (core.HeaderI, error) {
	panic("not implemented error")
}

// errCantSetPath returns an error if the path doesn't set properly
func (c *Chain) errCantSetPath(err error) error {
	return fmt.Errorf("path on chain %s failed to set: %w", c.ChainID(), err)
}

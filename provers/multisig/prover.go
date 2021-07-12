package multisig

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/modules/core/03-connection/types"
	chantypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/modules/core/exported"
	"github.com/hyperledger-labs/yui-relayer/core"
)

type LocalMultisigProver struct {
	config ClientConfig
	chain  core.ChainI
}

var _ core.ProverI = (*LocalMultisigProver)(nil)

func NewLocalMultisigProver(config ClientConfig, chain core.ChainI) *LocalMultisigProver {
	return &LocalMultisigProver{config: config, chain: chain}
}

func (pv *LocalMultisigProver) GetChainID() string {
	return pv.chain.ChainID()
}

func (pv *LocalMultisigProver) QueryClientConsensusStateWithProof(height int64, dstClientConsHeight ibcexported.Height) (*clienttypes.QueryConsensusStateResponse, error) {
	res, err := pv.chain.QueryClientConsensusState(height, dstClientConsHeight)
	if err != nil {
		return nil, err
	}
	_ = res
	panic("not implemented") // TODO: Implement
}

func (pv *LocalMultisigProver) QueryClientStateWithProof(height int64) (*clienttypes.QueryClientStateResponse, error) {
	res, err := pv.chain.QueryClientState(height)
	if err != nil {
		return nil, err
	}
	_ = res
	panic("not implemented") // TODO: Implement
}

func (pv *LocalMultisigProver) QueryConnectionWithProof(height int64) (*conntypes.QueryConnectionResponse, error) {
	res, err := pv.chain.QueryConnection(height)
	if err != nil {
		return nil, err
	}
	_ = res
	panic("not implemented") // TODO: Implement
}

func (pv *LocalMultisigProver) QueryChannelWithProof(height int64) (chanRes *chantypes.QueryChannelResponse, err error) {
	res, err := pv.chain.QueryChannel(height)
	if err != nil {
		return nil, err
	}
	_ = res
	panic("not implemented") // TODO: Implement
}

// QueryLatestHeight queries the chain for the latest height and returns it
func (pv *LocalMultisigProver) QueryLatestHeight() (int64, error) {
	panic("not implemented") // TODO: Implement
}

// QueryLatestHeader returns the latest header from the chain
func (pv *LocalMultisigProver) QueryLatestHeader() (out core.HeaderI, err error) {
	panic("not implemented") // TODO: Implement
}

// GetLatestLightHeight uses the CLI utilities to pull the latest height from a given chain
func (pv *LocalMultisigProver) GetLatestLightHeight() (int64, error) {
	panic("not implemented") // TODO: Implement
}

// MakeMsgCreateClient creates a CreateClientMsg to this chain
func (pv *LocalMultisigProver) MakeMsgCreateClient(clientID string, dstHeader core.HeaderI, signer sdk.AccAddress) (sdk.Msg, error) {
	panic("not implemented") // TODO: Implement
}

// CreateTrustedHeader creates ...
func (pv *LocalMultisigProver) CreateTrustedHeader(dstChain core.LightClientIBCQueryier, srcHeader core.HeaderI) (core.HeaderI, error) {
	panic("not implemented") // TODO: Implement
}

func (pv *LocalMultisigProver) UpdateLightWithHeader() (core.HeaderI, uint64, error) {
	panic("not implemented") // TODO: Implement
}

var _ core.ClientConfigI = (*ClientConfig)(nil)

func (cfg ClientConfig) BuildClient(chain core.ChainI) (core.ProverI, error) {
	prv := NewLocalMultisigProver(cfg, chain)
	return prv, nil
}

// GetPubKey unmarshals the public key into a cryptotypes.PubKey type.
// An error is returned if the public key is nil or the cached value
// is not a PubKey.
func (cs ClientConfig) GetPubKey() (cryptotypes.PubKey, error) {
	if cs.PublicKey == nil {
		return nil, sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "consensus state PublicKey cannot be nil")
	}

	publicKey, ok := cs.PublicKey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "consensus state PublicKey is not cryptotypes.PubKey")
	}

	return publicKey, nil
}

func (cs ClientConfig) MustGetPubKey() cryptotypes.PubKey {
	pubKey, err := cs.GetPubKey()
	if err != nil {
		panic(err)
	}
	return pubKey
}

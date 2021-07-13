package tendermint

import (
	"context"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/modules/core/03-connection/types"
	chantypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/modules/core/exported"
	ibcexported "github.com/cosmos/ibc-go/modules/core/exported"
	tmclient "github.com/cosmos/ibc-go/modules/light-clients/07-tendermint/types"
	"github.com/tendermint/tendermint/light"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/hyperledger-labs/yui-relayer/core"
)

type Prover struct {
	chain  *Chain
	config ClientConfig
}

var _ core.ProverI = (*Prover)(nil)

func NewProver(chain *Chain, config ClientConfig) *Prover {
	return &Prover{chain: chain, config: config}
}

func (pr *Prover) GetChainID() string {
	return pr.chain.ChainID()
}

func (pr *Prover) QueryClientConsensusStateWithProof(height int64, dstClientConsHeight ibcexported.Height) (*clienttypes.QueryConsensusStateResponse, error) {
	return pr.chain.queryClientConsensusState(height, dstClientConsHeight, true)
}

func (pr *Prover) QueryClientStateWithProof(height int64) (*clienttypes.QueryClientStateResponse, error) {
	return pr.chain.queryClientState(height, true)
}

func (pr *Prover) QueryConnectionWithProof(height int64) (*conntypes.QueryConnectionResponse, error) {
	return pr.chain.queryConnection(height, true)
}

func (pr *Prover) QueryChannelWithProof(height int64) (chanRes *chantypes.QueryChannelResponse, err error) {
	return pr.chain.queryChannel(height, true)
}

func (pr *Prover) QueryPacketCommitmentWithProof(height int64, seq uint64) (comRes *chantypes.QueryPacketCommitmentResponse, err error) {
	return pr.chain.queryPacketCommitment(height, seq, true)
}

func (pr *Prover) QueryPacketAcknowledgementCommitmentWithProof(height int64, seq uint64) (ackRes *chantypes.QueryPacketAcknowledgementResponse, err error) {
	return pr.chain.queryPacketAcknowledgementCommitment(height, seq, true)
}

// QueryLatestHeight queries the chain for the latest height and returns it
func (pr *Prover) QueryLatestHeight() (int64, error) {
	res, err := pr.chain.Client.Status(context.Background())
	if err != nil {
		return -1, err
	} else if res.SyncInfo.CatchingUp {
		return -1, fmt.Errorf("node at %s running chain %s not caught up", pr.chain.config.RpcAddr, pr.chain.ChainID())
	}

	return res.SyncInfo.LatestBlockHeight, nil
}

// QueryLatestHeader returns the latest header from the chain
func (pr *Prover) QueryLatestHeader() (out core.HeaderI, err error) {
	var h int64
	if h, err = pr.QueryLatestHeight(); err != nil {
		return nil, err
	}
	return pr.queryHeaderAtHeight(h)
}

// GetLatestLightHeight uses the CLI utilities to pull the latest height from a given chain
func (pr *Prover) GetLatestLightHeight() (int64, error) {
	db, df, err := pr.NewLightDB()
	if err != nil {
		return -1, err
	}
	defer df()

	client, err := pr.LightClient(db)
	if err != nil {
		return -1, err
	}

	return client.LastTrustedHeight()
}

// MakeMsgCreateClient creates a CreateClientMsg to this chain
func (pr *Prover) MakeMsgCreateClient(clientID string, dstHeader core.HeaderI, signer sdk.AccAddress) (sdk.Msg, error) {
	ubdPeriod, err := pr.chain.QueryUnbondingPeriod()
	if err != nil {
		return nil, err
	}
	consensusParams, err := pr.chain.QueryConsensusParams()
	if err != nil {
		return nil, err
	}
	return createClient(
		dstHeader.(*tmclient.Header),
		pr.getTrustingPeriod(),
		ubdPeriod,
		consensusParams,
		signer,
	), nil
}

// CreateTrustedHeader creates ...
func (pr *Prover) CreateTrustedHeader(dstChain core.LightClientIBCQueryier, srcHeader core.HeaderI) (core.HeaderI, error) {
	srcChain := pr.chain
	// make copy of header stored in mop
	tmp := srcHeader.(*tmclient.Header)
	h := *tmp

	dsth, err := dstChain.GetLatestLightHeight()
	if err != nil {
		return nil, err
	}

	// retrieve counterparty client from dst chain
	counterpartyClientRes, err := dstChain.QueryClientState(dsth)
	if err != nil {
		return nil, err
	}

	var cs exported.ClientState
	if err := srcChain.Encoding.Marshaler.UnpackAny(counterpartyClientRes.ClientState, &cs); err != nil {
		return nil, err
	}

	// inject TrustedHeight as latest height stored on counterparty client
	h.TrustedHeight = cs.GetLatestHeight().(clienttypes.Height)

	// query TrustedValidators at Trusted Height from srcChain
	valSet, err := srcChain.QueryValsetAtHeight(h.TrustedHeight)
	if err != nil {
		return nil, err
	}

	// inject TrustedValidators into header
	h.TrustedValidators = valSet
	return &h, nil
}

func lightError(err error) error { return fmt.Errorf("light client: %w", err) }

// UpdateLightWithHeader calls client.Update and then .
func (pr *Prover) UpdateLightWithHeader() (core.HeaderI, uint64, error) {
	// create database connection
	db, df, err := pr.NewLightDB()
	if err != nil {
		return nil, 0, lightError(err)
	}
	defer df()

	client, err := pr.LightClient(db)
	if err != nil {
		return nil, 0, lightError(err)
	}

	sh, err := client.Update(context.Background(), time.Now())
	if err != nil {
		return nil, 0, lightError(err)
	}

	if sh == nil {
		sh, err = client.TrustedLightBlock(0)
		if err != nil {
			return nil, 0, lightError(err)
		}
	}

	protoVal, err := tmtypes.NewValidatorSet(sh.ValidatorSet.Validators).ToProto()
	if err != nil {
		return nil, 0, err
	}

	h := &tmclient.Header{
		SignedHeader: sh.SignedHeader.ToProto(),
		ValidatorSet: protoVal,
	}
	return h, h.GetHeight().GetRevisionHeight(), nil
}

/// internal method ///

// getTrustingPeriod returns the trusting period for the chain
func (pr *Prover) getTrustingPeriod() time.Duration {
	tp, _ := time.ParseDuration(pr.config.TrustingPeriod)
	return tp
}

// queryHeaderAtHeight returns the header at a given height
func (c *Prover) queryHeaderAtHeight(height int64) (*tmclient.Header, error) {
	var (
		page    int = 1
		perPage int = 100000
	)
	if height <= 0 {
		return nil, fmt.Errorf("must pass in valid height, %d not valid", height)
	}

	res, err := c.chain.Client.Commit(context.Background(), &height)
	if err != nil {
		return nil, err
	}

	val, err := c.chain.Client.Validators(context.Background(), &height, &page, &perPage)
	if err != nil {
		return nil, err
	}

	protoVal, err := tmtypes.NewValidatorSet(val.Validators).ToProto()
	if err != nil {
		return nil, err
	}

	return &tmclient.Header{
		// NOTE: This is not a SignedHeader
		// We are missing a light.Commit type here
		SignedHeader: res.SignedHeader.ToProto(),
		ValidatorSet: protoVal,
	}, nil
}

// TrustOptions returns light.TrustOptions given a height and hash
func (pr *Prover) TrustOptions(height int64, hash []byte) light.TrustOptions {
	return light.TrustOptions{
		Period: pr.getTrustingPeriod(),
		Height: height,
		Hash:   hash,
	}
}

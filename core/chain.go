package core

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	conntypes "github.com/cosmos/ibc-go/modules/core/03-connection/types"
	chantypes "github.com/cosmos/ibc-go/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/modules/core/exported"
	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger-labs/yui-relayer/utils"
)

type Chain struct {
	ChainI
	ProverI
}

func NewChain(base ChainI, prover ProverI) *Chain {
	return &Chain{ChainI: base, ProverI: prover}
}

type ChainI interface {
	IBCQuerier

	ChainID() string
	ClientID() string

	GetAddress() (sdk.AccAddress, error)
	Marshaler() codec.Codec

	SetPath(p *PathEnd) error
	Path() *PathEnd

	SendMsgs(msgs []sdk.Msg) ([]byte, error)
	// Send sends msgs to the chain and logging a result of it
	// It returns a boolean value whether the result is success
	Send(msgs []sdk.Msg) bool

	StartEventListener(dst ChainI, strategy StrategyI)
	Init(homePath string, timeout time.Duration, debug bool) error
}

type IBCQuerier interface {
	// QueryClientConsensusState retrevies the latest consensus state for a client in state at a given height
	QueryClientConsensusState(height int64, dstClientConsHeight ibcexported.Height) (*clienttypes.QueryConsensusStateResponse, error)

	// QueryClientState returns the client state of dst chain
	// height represents the height of dst chain
	QueryClientState(height int64) (*clienttypes.QueryClientStateResponse, error)

	// QueryConnection returns the remote end of a given connection
	QueryConnection(height int64) (*conntypes.QueryConnectionResponse, error)

	// QueryChannel returns the channel associated with a channelID
	QueryChannel(height int64) (chanRes *chantypes.QueryChannelResponse, err error)

	// QueryBalance returns the amount of coins in the relayer account
	QueryBalance(address sdk.AccAddress) (sdk.Coins, error)
	// QueryDenomTraces returns all the denom traces from a given chain
	QueryDenomTraces(offset, limit uint64, height int64) (*transfertypes.QueryDenomTracesResponse, error)
	// QueryPacketCommitment returns the packet commitment proof at a given height
	QueryPacketCommitment(height int64, seq uint64) (comRes *chantypes.QueryPacketCommitmentResponse, err error)
	// QueryPacketCommitments returns an array of packet commitments
	QueryPacketCommitments(offset, limit, height uint64) (comRes *chantypes.QueryPacketCommitmentsResponse, err error)
	// QueryUnrecievedPackets returns a list of unrelayed packet commitments
	QueryUnrecievedPackets(height uint64, seqs []uint64) ([]uint64, error)
	// QueryPacketAcknowledgements returns an array of packet acks
	QueryPacketAcknowledgements(offset, limit, height uint64) (comRes *chantypes.QueryPacketAcknowledgementsResponse, err error)
	// QueryUnrecievedAcknowledgements returns a list of unrelayed packet acks
	QueryUnrecievedAcknowledgements(height uint64, seqs []uint64) ([]uint64, error)
	// QueryPacketAcknowledgementCommitment returns the packet ack proof at a given height
	QueryPacketAcknowledgementCommitment(height int64, seq uint64) (ackRes *chantypes.QueryPacketAcknowledgementResponse, err error)

	// QueryPacket returns a packet corresponds to a given sequence
	QueryPacket(height int64, sequence uint64) (*chantypes.Packet, error)
	QueryPacketAcknowledgement(height int64, sequence uint64) ([]byte, error)
}

type IBCProvableQuerier interface {
	QueryClientConsensusStateWithProof(height int64, dstClientConsHeight ibcexported.Height) (*clienttypes.QueryConsensusStateResponse, error)
	QueryClientStateWithProof(height int64) (*clienttypes.QueryClientStateResponse, error)
	QueryConnectionWithProof(height int64) (*conntypes.QueryConnectionResponse, error)
	QueryChannelWithProof(height int64) (chanRes *chantypes.QueryChannelResponse, err error)
	QueryPacketCommitmentWithProof(height int64, seq uint64) (comRes *chantypes.QueryPacketCommitmentResponse, err error)
	QueryPacketAcknowledgementCommitmentWithProof(height int64, seq uint64) (ackRes *chantypes.QueryPacketAcknowledgementResponse, err error)
}

type LightClientIBCQueryier interface {
	LightClientI
	IBCQuerier
}

type ChainClientConfig struct {
	Chain  json.RawMessage `json:"chain" yaml:"chain"` // NOTE: it's any type as json format
	Client json.RawMessage `json:"client" yaml:"client"`

	// cache
	chain  ChainConfigI  `json:"-" yaml:"-"`
	client ClientConfigI `json:"-" yaml:"-"`
}

func NewChainClientConfig(m codec.JSONCodec, chain ChainConfigI, client ClientConfigI) (*ChainClientConfig, error) {
	cbz, err := utils.MarshalJSONAny(m, chain)
	if err != nil {
		return nil, err
	}
	clbz, err := utils.MarshalJSONAny(m, client)
	if err != nil {
		return nil, err
	}
	return &ChainClientConfig{
		Chain:  cbz,
		Client: clbz,
		chain:  chain,
		client: client,
	}, nil
}

func (cc *ChainClientConfig) Init(m codec.Codec) error {
	var chain ChainConfigI
	if err := utils.UnmarshalJSONAny(m, &chain, cc.Chain); err != nil {
		return err
	}
	var client ClientConfigI
	if err := utils.UnmarshalJSONAny(m, &client, cc.Client); err != nil {
		return err
	}
	cc.chain = chain
	cc.client = client
	return nil
}

func (cc ChainClientConfig) GetChainConfig() (ChainConfigI, error) {
	if cc.chain == nil {
		return nil, errors.New("chain is nil")
	}
	return cc.chain, nil
}

func (cc ChainClientConfig) GetClientConfig() (ClientConfigI, error) {
	if cc.client == nil {
		return nil, errors.New("client is nil")
	}
	return cc.client, nil
}

func (cc ChainClientConfig) BuildChain() (*Chain, error) {
	chainConfig, err := cc.GetChainConfig()
	if err != nil {
		return nil, err
	}
	clientConfig, err := cc.GetClientConfig()
	if err != nil {
		return nil, err
	}
	chain, err := chainConfig.BuildChain()
	if err != nil {
		return nil, err
	}
	client, err := clientConfig.BuildClient(chain)
	if err != nil {
		return nil, err
	}
	return &Chain{ChainI: chain, ProverI: client}, nil
}

type ChainConfigI interface {
	proto.Message
	BuildChain() (ChainI, error)
}

type ClientConfigI interface {
	proto.Message
	BuildClient(ChainI) (ProverI, error)
}

package core

import sdk "github.com/cosmos/cosmos-sdk/types"

type ProverI interface {
	IBCProvableQuerier
	LightClientI
}

type LightClientI interface {
	// QueryLatestHeight queries the chain for the latest height and returns it
	QueryLatestHeight() (int64, error)
	// QueryLatestHeader returns the latest header from the chain
	QueryLatestHeader() (out HeaderI, err error)

	// GetLatestLightHeight uses the CLI utilities to pull the latest height from a given chain
	GetLatestLightHeight() (int64, error)

	// MakeMsgCreateClient creates a CreateClientMsg to this chain
	MakeMsgCreateClient(clientID string, dstHeader HeaderI, signer sdk.AccAddress) (sdk.Msg, error)

	// CreateTrustedHeader creates ...
	CreateTrustedHeader(dstChain LightClientIBCQueryier, srcHeader HeaderI) (HeaderI, error)
	UpdateLightWithHeader() (HeaderI, error)
}

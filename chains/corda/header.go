package corda

import (
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	ibcexported "github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	"github.com/datachainlab/relayer/core"
)

type CordaHeader struct{}

var _ core.HeaderI = (*CordaHeader)(nil)

func (*CordaHeader) ClientType() string {
	return "corda"
}

func (*CordaHeader) GetHeight() ibcexported.Height {
	return clienttypes.Height{}
}

func (*CordaHeader) ValidateBasic() error {
	return nil
}

package multisig

// import (
// 	"context"

// 	"github.com/hyperledger-labs/yui-relayer/core"
// 	"github.com/hyperledger-labs/yui-relayer/provers"
// 	"google.golang.org/grpc"
// )

// type LocalMultisigProver struct {
// 	config ClientConfig
// }

// var _ provers.ProverClient = (*LocalMultisigProver)(nil)

// func NewLocalMultisigProver(config ClientConfig) *LocalMultisigProver {
// 	return &LocalMultisigProver{config: config}
// }

// func (prv *LocalMultisigProver) Prove(ctx context.Context, in *provers.QueryProveRequest, _ ...grpc.CallOption) (*provers.QueryProveResponse, error) {
// 	// TODO implements
// 	return nil, nil
// }

// var _ core.ExternalProverConfigI = (*ClientConfig)(nil)

// func (ClientConfig) External() {}

// func (cfg ClientConfig) BuildProver() (provers.ProverClient, error) {
// 	prv := NewLocalMultisigProver(cfg)
// 	return prv, nil
// }

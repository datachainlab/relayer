package core

import (
	"fmt"

	"github.com/cosmos/ibc-go/modules/core/exported"
)

type HeaderI interface {
	exported.Header
}

type SyncHeadersI interface {
	// GetChainHeight returns the height of chain
	GetChainHeight(chainID string) uint64
	// GetLightHeader returns the latest header of light client
	GetLightHeader(chainID string) HeaderI
	GetNextLightHeader(src, dst LightClientIBCQueryier) (HeaderI, error)
	GetNextLightHeaders(src, dst LightClientIBCQueryier) (srcHeader HeaderI, dstHeader HeaderI, err error)
	Updates(LightClientI, LightClientI) error
}

type syncHeaders struct {
	latestLightHeaders map[string]HeaderI // chainID => HeaderI
	latestChainHeights map[string]uint64  // chainID => height
}

var _ SyncHeadersI = (*syncHeaders)(nil)

// NewSyncHeaders returns a new instance of SyncHeadersI that can be easily
// kept "reasonably up to date"
func NewSyncHeaders(src, dst LightClientI) (SyncHeadersI, error) {
	sh := &syncHeaders{
		latestLightHeaders: map[string]HeaderI{src.GetChainID(): nil, dst.GetChainID(): nil},
		latestChainHeights: map[string]uint64{src.GetChainID(): 0, dst.GetChainID(): 0},
	}
	if err := sh.Updates(src, dst); err != nil {
		return nil, err
	}
	return sh, nil
}

func (sh syncHeaders) GetChainHeight(chainID string) uint64 {
	// return sh.latestLightHeaders[chainID].GetHeight().GetRevisionHeight()
	return sh.latestChainHeights[chainID]
}

func (sh syncHeaders) GetLightHeader(chainID string) HeaderI {
	return sh.latestLightHeaders[chainID]
}

func (sh syncHeaders) GetNextLightHeader(src, dst LightClientIBCQueryier) (HeaderI, error) {
	return src.CreateTrustedHeader(dst, sh.GetLightHeader(src.GetChainID()))
}

func (sh syncHeaders) GetNextLightHeaders(src, dst LightClientIBCQueryier) (HeaderI, HeaderI, error) {
	srcTh, err := sh.GetNextLightHeader(src, dst)
	if err != nil {
		fmt.Println("failed to GetTrustedHeaders(src):", err)
		return nil, nil, err
	}
	dstTh, err := sh.GetNextLightHeader(dst, src)
	if err != nil {
		fmt.Println("failed to GetTrustedHeaders(dst):", err)
		return nil, nil, err
	}
	return srcTh, dstTh, nil
}

func (sh *syncHeaders) Updates(src, dst LightClientI) error {
	srcHeader, srcHeight, err := src.UpdateLightWithHeader()
	if err != nil {
		return err
	}
	dstHeader, dstHeight, err := dst.UpdateLightWithHeader()
	if err != nil {
		return err
	}

	sh.latestLightHeaders[src.GetChainID()] = srcHeader
	sh.latestLightHeaders[dst.GetChainID()] = dstHeader

	sh.latestChainHeights[src.GetChainID()] = srcHeight
	sh.latestChainHeights[dst.GetChainID()] = dstHeight
	return nil
}

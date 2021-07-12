package multisig

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/modules/core/24-host"
	"github.com/cosmos/ibc-go/modules/core/exported"
	solomachinetypes "github.com/cosmos/ibc-go/modules/light-clients/06-solomachine/types"
)

// NOTE: This codebase is based on cosmos/ibc-go/testing/solomachine.go

// Solomachine is a testing helper used to simulate a counterparty
// solo machine client.
type Solomachine struct {
	cdc         codec.BinaryCodec
	ClientID    string
	PrivateKeys []cryptotypes.PrivKey // keys used for signing
	PublicKeys  []cryptotypes.PubKey  // keys used for generating solo machine pub key
	PublicKey   cryptotypes.PubKey    // key used for verification
	Sequence    uint64
	Time        uint64
	Diversifier string
	Prefix      commitmenttypes.MerklePrefix
}

// NewSolomachine returns a new solomachine instance with an `nKeys` amount of
// generated private/public key pairs and a sequence starting at 1. If nKeys
// is greater than 1 then a multisig public key is used.
func NewSolomachine(cdc codec.BinaryCodec, clientID, diversifier string, nKeys uint64) (*Solomachine, error) {
	privKeys, pubKeys, pk, err := GenerateKeys(nKeys)
	if err != nil {
		return nil, err
	}

	return &Solomachine{
		cdc:         cdc,
		ClientID:    clientID,
		PrivateKeys: privKeys,
		PublicKeys:  pubKeys,
		PublicKey:   pk,
		Sequence:    1,
		Time:        10,
		Diversifier: diversifier,
		Prefix:      commitmenttypes.NewMerklePrefix([]byte("ibc")),
	}, nil
}

// GenerateKeys generates a new set of secp256k1 private keys and public keys.
// If the number of keys is greater than one then the public key returned represents
// a multisig public key. The private keys are used for signing, the public
// keys are used for generating the public key and the public key is used for
// solo machine verification. The usage of secp256k1 is entirely arbitrary.
// The key type can be swapped for any key type supported by the PublicKey
// interface, if needed. The same is true for the amino based Multisignature
// public key.
func GenerateKeys(n uint64) ([]cryptotypes.PrivKey, []cryptotypes.PubKey, cryptotypes.PubKey, error) {
	if n == 0 {
		return nil, nil, nil, errors.New("generation of zero keys is not allowed")
	}

	privKeys := make([]cryptotypes.PrivKey, n)
	pubKeys := make([]cryptotypes.PubKey, n)
	for i := uint64(0); i < n; i++ {
		privKeys[i] = secp256k1.GenPrivKey()
		pubKeys[i] = privKeys[i].PubKey()
	}

	var pk cryptotypes.PubKey
	if len(privKeys) > 1 {
		// generate multi sig pk
		pk = kmultisig.NewLegacyAminoPubKey(int(n), pubKeys)
	} else {
		pk = privKeys[0].PubKey()
	}

	return privKeys, pubKeys, pk, nil
}

// ClientState returns a new solo machine ClientState instance. Default usage does not allow update
// after governance proposal
func (solo *Solomachine) ClientState() *solomachinetypes.ClientState {
	return solomachinetypes.NewClientState(solo.Sequence, solo.ConsensusState(), false)
}

// ConsensusState returns a new solo machine ConsensusState instance
func (solo *Solomachine) ConsensusState() *solomachinetypes.ConsensusState {
	publicKey, err := codectypes.NewAnyWithValue(solo.PublicKey)
	if err != nil {
		panic(err)
	}

	return &solomachinetypes.ConsensusState{
		PublicKey:   publicKey,
		Diversifier: solo.Diversifier,
		Timestamp:   solo.Time,
	}
}

// GetHeight returns an exported.Height with Sequence as RevisionHeight
func (solo *Solomachine) GetHeight() exported.Height {
	return clienttypes.NewHeight(0, solo.Sequence)
}

// CreateHeader generates a new private/public key pair and creates the
// necessary signature to construct a valid solo machine header.
func (solo *Solomachine) CreateHeader() (*solomachinetypes.Header, error) {
	// generate new private keys and signature for header
	newPrivKeys, newPubKeys, newPubKey, err := GenerateKeys(uint64(len(solo.PrivateKeys)))
	if err != nil {
		return nil, err
	}

	publicKey, err := codectypes.NewAnyWithValue(newPubKey)
	if err != nil {
		return nil, err
	}

	data := &solomachinetypes.HeaderData{
		NewPubKey:      publicKey,
		NewDiversifier: solo.Diversifier,
	}

	dataBz, err := solo.cdc.Marshal(data)
	if err != nil {
		return nil, err
	}

	signBytes := &solomachinetypes.SignBytes{
		Sequence:    solo.Sequence,
		Timestamp:   solo.Time,
		Diversifier: solo.Diversifier,
		DataType:    solomachinetypes.HEADER,
		Data:        dataBz,
	}

	bz, err := solo.cdc.Marshal(signBytes)
	if err != nil {
		return nil, err
	}

	sig, err := solo.GenerateSignature(bz)
	if err != nil {
		return nil, err
	}

	header := &solomachinetypes.Header{
		Sequence:       solo.Sequence,
		Timestamp:      solo.Time,
		Signature:      sig,
		NewPublicKey:   publicKey,
		NewDiversifier: solo.Diversifier,
	}

	// assumes successful header update
	solo.Sequence++
	solo.PrivateKeys = newPrivKeys
	solo.PublicKeys = newPubKeys
	solo.PublicKey = newPubKey

	return header, nil
}

// // CreateMisbehaviour constructs testing misbehaviour for the solo machine client
// // by signing over two different data bytes at the same sequence.
// func (solo *Solomachine) CreateMisbehaviour() *solomachinetypes.Misbehaviour {
// 	path := solo.GetClientStatePath("counterparty")
// 	dataOne, err := solomachinetypes.ClientStateDataBytes(solo.cdc, path, solo.ClientState())
// 	require.NoError(solo.t, err)

// 	path = solo.GetConsensusStatePath("counterparty", clienttypes.NewHeight(0, 1))
// 	dataTwo, err := solomachinetypes.ConsensusStateDataBytes(solo.cdc, path, solo.ConsensusState())
// 	require.NoError(solo.t, err)

// 	signBytes := &solomachinetypes.SignBytes{
// 		Sequence:    solo.Sequence,
// 		Timestamp:   solo.Time,
// 		Diversifier: solo.Diversifier,
// 		DataType:    solomachinetypes.CLIENT,
// 		Data:        dataOne,
// 	}

// 	bz, err := solo.cdc.Marshal(signBytes)
// 	require.NoError(solo.t, err)

// 	sig := solo.GenerateSignature(bz)
// 	signatureOne := solomachinetypes.SignatureAndData{
// 		Signature: sig,
// 		DataType:  solomachinetypes.CLIENT,
// 		Data:      dataOne,
// 		Timestamp: solo.Time,
// 	}

// 	// misbehaviour signaturess can have different timestamps
// 	solo.Time++

// 	signBytes = &solomachinetypes.SignBytes{
// 		Sequence:    solo.Sequence,
// 		Timestamp:   solo.Time,
// 		Diversifier: solo.Diversifier,
// 		DataType:    solomachinetypes.CONSENSUS,
// 		Data:        dataTwo,
// 	}

// 	bz, err = solo.cdc.Marshal(signBytes)
// 	require.NoError(solo.t, err)

// 	sig = solo.GenerateSignature(bz)
// 	signatureTwo := solomachinetypes.SignatureAndData{
// 		Signature: sig,
// 		DataType:  solomachinetypes.CONSENSUS,
// 		Data:      dataTwo,
// 		Timestamp: solo.Time,
// 	}

// 	return &solomachinetypes.Misbehaviour{
// 		ClientId:     solo.ClientID,
// 		Sequence:     solo.Sequence,
// 		SignatureOne: &signatureOne,
// 		SignatureTwo: &signatureTwo,
// 	}
// }

// GenerateSignature uses the stored private keys to generate a signature
// over the sign bytes with each key. If the amount of keys is greater than
// 1 then a multisig data type is returned.
func (solo *Solomachine) GenerateSignature(signBytes []byte) ([]byte, error) {
	sigs := make([]signing.SignatureData, len(solo.PrivateKeys))
	for i, key := range solo.PrivateKeys {
		sig, err := key.Sign(signBytes)
		if err != nil {
			return nil, err
		}
		sigs[i] = &signing.SingleSignatureData{
			Signature: sig,
		}
	}

	var sigData signing.SignatureData
	if len(sigs) == 1 {
		// single public key
		sigData = sigs[0]
	} else {
		// generate multi signature data
		multiSigData := multisig.NewMultisig(len(sigs))
		for i, sig := range sigs {
			multisig.AddSignature(multiSigData, sig, i)
		}

		sigData = multiSigData
	}

	protoSigData := signing.SignatureDataToProto(sigData)
	return solo.cdc.Marshal(protoSigData)
}

// GetClientStatePath returns the commitment path for the client state.
func (solo *Solomachine) GetClientStatePath(counterpartyClientIdentifier string) commitmenttypes.MerklePath {
	path, err := commitmenttypes.ApplyPrefix(solo.Prefix, commitmenttypes.NewMerklePath(host.FullClientStatePath(counterpartyClientIdentifier)))
	if err != nil {
		panic(err)
	}
	return path
}

// GetConsensusStatePath returns the commitment path for the consensus state.
func (solo *Solomachine) GetConsensusStatePath(counterpartyClientIdentifier string, consensusHeight exported.Height) commitmenttypes.MerklePath {
	path, err := commitmenttypes.ApplyPrefix(solo.Prefix, commitmenttypes.NewMerklePath(host.FullConsensusStatePath(counterpartyClientIdentifier, consensusHeight)))
	if err != nil {
		panic(err)
	}
	return path
}

// GetConnectionStatePath returns the commitment path for the connection state.
func (solo *Solomachine) GetConnectionStatePath(connID string) commitmenttypes.MerklePath {
	connectionPath := commitmenttypes.NewMerklePath(host.ConnectionPath(connID))
	path, err := commitmenttypes.ApplyPrefix(solo.Prefix, connectionPath)
	if err != nil {
		panic(err)
	}
	return path
}

// GetChannelStatePath returns the commitment path for that channel state.
func (solo *Solomachine) GetChannelStatePath(portID, channelID string) commitmenttypes.MerklePath {
	channelPath := commitmenttypes.NewMerklePath(host.ChannelPath(portID, channelID))
	path, err := commitmenttypes.ApplyPrefix(solo.Prefix, channelPath)
	if err != nil {
		panic(err)
	}
	return path
}

// GetPacketCommitmentPath returns the commitment path for a packet commitment.
func (solo *Solomachine) GetPacketCommitmentPath(portID, channelID string) commitmenttypes.MerklePath {
	commitmentPath := commitmenttypes.NewMerklePath(host.PacketCommitmentPath(portID, channelID, solo.Sequence))
	path, err := commitmenttypes.ApplyPrefix(solo.Prefix, commitmentPath)
	if err != nil {
		panic(err)
	}
	return path
}

// GetPacketAcknowledgementPath returns the commitment path for a packet acknowledgement.
func (solo *Solomachine) GetPacketAcknowledgementPath(portID, channelID string) commitmenttypes.MerklePath {
	ackPath := commitmenttypes.NewMerklePath(host.PacketAcknowledgementPath(portID, channelID, solo.Sequence))
	path, err := commitmenttypes.ApplyPrefix(solo.Prefix, ackPath)
	if err != nil {
		panic(err)
	}
	return path
}

// GetPacketReceiptPath returns the commitment path for a packet receipt
// and an absent receipts.
func (solo *Solomachine) GetPacketReceiptPath(portID, channelID string) commitmenttypes.MerklePath {
	receiptPath := commitmenttypes.NewMerklePath(host.PacketReceiptPath(portID, channelID, solo.Sequence))
	path, err := commitmenttypes.ApplyPrefix(solo.Prefix, receiptPath)
	if err != nil {
		panic(err)
	}
	return path
}

// GetNextSequenceRecvPath returns the commitment path for the next sequence recv counter.
func (solo *Solomachine) GetNextSequenceRecvPath(portID, channelID string) commitmenttypes.MerklePath {
	nextSequenceRecvPath := commitmenttypes.NewMerklePath(host.NextSequenceRecvPath(portID, channelID))
	path, err := commitmenttypes.ApplyPrefix(solo.Prefix, nextSequenceRecvPath)
	if err != nil {
		panic(err)
	}
	return path
}

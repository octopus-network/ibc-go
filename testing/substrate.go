package ibctesting

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/stretchr/testify/require"

	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v3/modules/core/23-commitment/types"
	host "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	solomachinetypes "github.com/cosmos/ibc-go/v3/modules/light-clients/06-solomachine/types"
	//TODO: import substrate types 
)

// Substrate is a testing helper used to simulate a counterparty
// substrate solo machine client.
type Substrate struct {
	t *testing.T

	cdc         codec.BinaryCodec
	ClientID    string
	PrivateKeys []cryptotypes.PrivKey // keys used for signing
	PublicKeys  []cryptotypes.PubKey  // keys used for generating solo machine pub key
	PublicKey   cryptotypes.PubKey    // key used for verification
	Sequence    uint64
	Time        uint64
	Diversifier string
}

// NewSolomachine returns a new solomachine instance with an `nKeys` amount of
// generated private/public key pairs and a sequence starting at 1. If nKeys
// is greater than 1 then a multisig public key is used.
func NewSubstrate(t *testing.T, cdc codec.BinaryCodec, clientID, diversifier string, nKeys uint64) *Substrate {
	privKeys, pubKeys, pk := GenerateKeys(t, nKeys)

	return &Substrate{
		t:           t,
		cdc:         cdc,
		ClientID:    clientID,
		PrivateKeys: privKeys,
		PublicKeys:  pubKeys,
		PublicKey:   pk,
		Sequence:    1,
		Time:        10,
		Diversifier: diversifier,
	}
}

// GenerateKeys generates a new set of secp256k1 private keys and public keys.
// If the number of keys is greater than one then the public key returned represents
// a multisig public key. The private keys are used for signing, the public
// keys are used for generating the public key and the public key is used for
// solo machine verification. The usage of secp256k1 is entirely arbitrary.
// The key type can be swapped for any key type supported by the PublicKey
// interface, if needed. The same is true for the amino based Multisignature
// public key.
func GenerateSubKeys(t *testing.T, n uint64) ([]cryptotypes.PrivKey, []cryptotypes.PubKey, cryptotypes.PubKey) {
	require.NotEqual(t, uint64(0), n, "generation of zero keys is not allowed")

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

	return privKeys, pubKeys, pk
}

// ClientState returns a new solo machine ClientState instance. Default usage does not allow update
// after governance proposal
func (sub *Substrate) ClientState() *solomachinetypes.ClientState {
	return solomachinetypes.NewClientState(sub.Sequence, sub.ConsensusState(), false)
}

// ConsensusState returns a new solo machine ConsensusState instance
func (sub *Substrate) ConsensusState() *solomachinetypes.ConsensusState {
	publicKey, err := codectypes.NewAnyWithValue(sub.PublicKey)
	require.NoError(sub.t, err)

	return &solomachinetypes.ConsensusState{
		PublicKey:   publicKey,
		Diversifier: sub.Diversifier,
		Timestamp:   sub.Time,
	}
}

// GetHeight returns an exported.Height with Sequence as RevisionHeight
func (sub *Substrate) GetHeight() exported.Height {
	return clienttypes.NewHeight(0, sub.Sequence)
}

// CreateHeader generates a new private/public key pair and creates the
// necessary signature to construct a valid solo machine header.
func (sub *Substrate) CreateHeader() *solomachinetypes.Header {
	// generate new private keys and signature for header
	newPrivKeys, newPubKeys, newPubKey := GenerateKeys(sub.t, uint64(len(sub.PrivateKeys)))

	publicKey, err := codectypes.NewAnyWithValue(newPubKey)
	require.NoError(sub.t, err)

	data := &solomachinetypes.HeaderData{
		NewPubKey:      publicKey,
		NewDiversifier: sub.Diversifier,
	}

	dataBz, err := sub.cdc.Marshal(data)
	require.NoError(sub.t, err)

	signBytes := &solomachinetypes.SignBytes{
		Sequence:    sub.Sequence,
		Timestamp:   sub.Time,
		Diversifier: sub.Diversifier,
		DataType:    solomachinetypes.HEADER,
		Data:        dataBz,
	}

	bz, err := sub.cdc.Marshal(signBytes)
	require.NoError(sub.t, err)

	sig := sub.GenerateSignature(bz)

	header := &solomachinetypes.Header{
		Sequence:       sub.Sequence,
		Timestamp:      sub.Time,
		Signature:      sig,
		NewPublicKey:   publicKey,
		NewDiversifier: sub.Diversifier,
	}

	// assumes successful header update
	sub.Sequence++
	sub.PrivateKeys = newPrivKeys
	sub.PublicKeys = newPubKeys
	sub.PublicKey = newPubKey

	return header
}

// CreateMisbehaviour constructs testing misbehaviour for the solo machine client
// by signing over two different data bytes at the same sequence.
func (sub *Substrate) CreateMisbehaviour() *solomachinetypes.Misbehaviour {
	path := sub.GetClientStatePath("counterparty")
	dataOne, err := solomachinetypes.ClientStateDataBytes(sub.cdc, path, sub.ClientState())
	require.NoError(sub.t, err)

	path = sub.GetConsensusStatePath("counterparty", clienttypes.NewHeight(0, 1))
	dataTwo, err := solomachinetypes.ConsensusStateDataBytes(sub.cdc, path, sub.ConsensusState())
	require.NoError(sub.t, err)

	signBytes := &solomachinetypes.SignBytes{
		Sequence:    sub.Sequence,
		Timestamp:   sub.Time,
		Diversifier: sub.Diversifier,
		DataType:    solomachinetypes.CLIENT,
		Data:        dataOne,
	}

	bz, err := sub.cdc.Marshal(signBytes)
	require.NoError(sub.t, err)

	sig := sub.GenerateSignature(bz)
	signatureOne := solomachinetypes.SignatureAndData{
		Signature: sig,
		DataType:  solomachinetypes.CLIENT,
		Data:      dataOne,
		Timestamp: sub.Time,
	}

	// misbehaviour signaturess can have different timestamps
	sub.Time++

	signBytes = &solomachinetypes.SignBytes{
		Sequence:    sub.Sequence,
		Timestamp:   sub.Time,
		Diversifier: sub.Diversifier,
		DataType:    solomachinetypes.CONSENSUS,
		Data:        dataTwo,
	}

	bz, err = sub.cdc.Marshal(signBytes)
	require.NoError(sub.t, err)

	sig = sub.GenerateSignature(bz)
	signatureTwo := solomachinetypes.SignatureAndData{
		Signature: sig,
		DataType:  solomachinetypes.CONSENSUS,
		Data:      dataTwo,
		Timestamp: sub.Time,
	}

	return &solomachinetypes.Misbehaviour{
		ClientId:     sub.ClientID,
		Sequence:     sub.Sequence,
		SignatureOne: &signatureOne,
		SignatureTwo: &signatureTwo,
	}
}

// GenerateSignature uses the stored private keys to generate a signature
// over the sign bytes with each key. If the amount of keys is greater than
// 1 then a multisig data type is returned.
func (sub *Substrate) GenerateSignature(signBytes []byte) []byte {
	sigs := make([]signing.SignatureData, len(sub.PrivateKeys))
	for i, key := range sub.PrivateKeys {
		sig, err := key.Sign(signBytes)
		require.NoError(sub.t, err)

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
	bz, err := sub.cdc.Marshal(protoSigData)
	require.NoError(sub.t, err)

	return bz
}

// GetClientStatePath returns the commitment path for the client state.
func (sub *Substrate) GetClientStatePath(counterpartyClientIdentifier string) commitmenttypes.MerklePath {
	path, err := commitmenttypes.ApplyPrefix(prefix, commitmenttypes.NewMerklePath(host.FullClientStatePath(counterpartyClientIdentifier)))
	require.NoError(sub.t, err)

	return path
}

// GetConsensusStatePath returns the commitment path for the consensus state.
func (sub *Substrate) GetConsensusStatePath(counterpartyClientIdentifier string, consensusHeight exported.Height) commitmenttypes.MerklePath {
	path, err := commitmenttypes.ApplyPrefix(prefix, commitmenttypes.NewMerklePath(host.FullConsensusStatePath(counterpartyClientIdentifier, consensusHeight)))
	require.NoError(sub.t, err)

	return path
}

// GetConnectionStatePath returns the commitment path for the connection state.
func (sub *Substrate) GetConnectionStatePath(connID string) commitmenttypes.MerklePath {
	connectionPath := commitmenttypes.NewMerklePath(host.ConnectionPath(connID))
	path, err := commitmenttypes.ApplyPrefix(prefix, connectionPath)
	require.NoError(sub.t, err)

	return path
}

// GetChannelStatePath returns the commitment path for that channel state.
func (sub *Substrate) GetChannelStatePath(portID, channelID string) commitmenttypes.MerklePath {
	channelPath := commitmenttypes.NewMerklePath(host.ChannelPath(portID, channelID))
	path, err := commitmenttypes.ApplyPrefix(prefix, channelPath)
	require.NoError(sub.t, err)

	return path
}

// GetPacketCommitmentPath returns the commitment path for a packet commitment.
func (sub *Substrate) GetPacketCommitmentPath(portID, channelID string) commitmenttypes.MerklePath {
	commitmentPath := commitmenttypes.NewMerklePath(host.PacketCommitmentPath(portID, channelID, sub.Sequence))
	path, err := commitmenttypes.ApplyPrefix(prefix, commitmentPath)
	require.NoError(sub.t, err)

	return path
}

// GetPacketAcknowledgementPath returns the commitment path for a packet acknowledgement.
func (sub *Substrate) GetPacketAcknowledgementPath(portID, channelID string) commitmenttypes.MerklePath {
	ackPath := commitmenttypes.NewMerklePath(host.PacketAcknowledgementPath(portID, channelID, sub.Sequence))
	path, err := commitmenttypes.ApplyPrefix(prefix, ackPath)
	require.NoError(sub.t, err)

	return path
}

// GetPacketReceiptPath returns the commitment path for a packet receipt
// and an absent receipts.
func (sub *Substrate) GetPacketReceiptPath(portID, channelID string) commitmenttypes.MerklePath {
	receiptPath := commitmenttypes.NewMerklePath(host.PacketReceiptPath(portID, channelID, sub.Sequence))
	path, err := commitmenttypes.ApplyPrefix(prefix, receiptPath)
	require.NoError(sub.t, err)

	return path
}

// GetNextSequenceRecvPath returns the commitment path for the next sequence recv counter.
func (sub *Substrate) GetNextSequenceRecvPath(portID, channelID string) commitmenttypes.MerklePath {
	nextSequenceRecvPath := commitmenttypes.NewMerklePath(host.NextSequenceRecvPath(portID, channelID))
	path, err := commitmenttypes.ApplyPrefix(prefix, nextSequenceRecvPath)
	require.NoError(sub.t, err)

	return path
}

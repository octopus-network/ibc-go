package types

import (
	time "time"

	clienttypes "github.com/cosmos/ibc-go/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/modules/core/exported"
)

var _ exported.Header = &Header{}

// ConsensusState returns the updated consensus state associated with the header
func (h Header) ConsensusState() *ConsensusState {

	return &ConsensusState{
		/// The parent hash.
		ParentHash: h.BlockHeader.ParentHash,
		/// The block number.
		BlockNumber: h.BlockHeader.BlockNumber,
		/// The state trie merkle root
		StateRoot: h.BlockHeader.StateRoot,
		/// The merkle root of the extrinsics.
		ExtrinsicsRoot: h.BlockHeader.ExtrinsicsRoot,
		/// A chain-specific digest of data useful for light clients or referencing auxiliary data.
		Digest:    h.BlockHeader.Digest,
		Root:      commitmenttypes.NewMerkleRoot([]byte(h.BlockHeader.ExtrinsicsRoot)),
		Timestamp: time.Unix(0, 0),
	}
}

// ClientType defines that the Header is a Tendermint consensus algorithm
func (h Header) ClientType() string {
	return exported.Grandpa
}

// GetHeight returns the current height. It returns 0 if the grandpa
// header is nil.
// NOTE: the header.Header is checked to be non nil in ValidateBasic.
func (h Header) GetHeight() exported.Height {

	return clienttypes.NewHeight(0, uint64(h.BlockHeader.BlockNumber))
}

// // GetTime returns the current block timestamp. It returns a zero time if
// // the tendermint header is nil.
// // NOTE: the header.Header is checked to be non nil in ValidateBasic.
func (h Header) GetTime() time.Time {
	return time.Unix(0, 0)
}

// ValidateBasic calls the SignedHeader ValidateBasic function and checks
// that validatorsets are not nil.
// NOTE: TrustedHeight and TrustedValidators may be empty when creating client
// with MsgCreateClient
func (h Header) ValidateBasic() error {
	// if h.SignedHeader == nil {
	// 	return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "tendermint signed header cannot be nil")
	// }
	// if h.Height == nil {
	// 	return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "Grandpa height cannot be nil")
	// }
	// tmSignedHeader, err := tmtypes.SignedHeaderFromProto(h.SignedHeader)
	// if err != nil {
	// 	return sdkerrors.Wrap(err, "header is not a tendermint header")
	// }
	// if err := tmSignedHeader.ValidateBasic(h.Header.GetChainID()); err != nil {
	// 	return sdkerrors.Wrap(err, "header failed basic validation")
	// }

	// TrustedHeight is less than Header for updates and misbehaviour
	// if h.TrustedHeight.GTE(h.GetHeight()) {
	// 	return sdkerrors.Wrapf(ErrInvalidHeaderHeight, "TrustedHeight %d must be less than header height %d",
	// 		h.TrustedHeight, h.GetHeight())
	// }

	// if h.ValidatorSet == nil {
	// 	return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "validator set is nil")
	// }
	// tmValset, err := tmtypes.ValidatorSetFromProto(h.ValidatorSet)
	// if err != nil {
	// 	return sdkerrors.Wrap(err, "validator set is not tendermint validator set")
	// }
	// if !bytes.Equal(h.Header.ValidatorsHash, tmValset.Hash()) {
	// 	return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "validator set does not match hash")
	// }
	return nil
}

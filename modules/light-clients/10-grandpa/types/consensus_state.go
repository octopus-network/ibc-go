package types

import (
	time "time"

	commitmenttypes "github.com/cosmos/ibc-go/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/modules/core/exported"
)

// SentinelRoot is used as a stand-in root value for the consensus state set at the upgrade height
const SentinelRoot = "sentinel_root"

// NewConsensusState creates a new ConsensusState instance.
func NewConsensusState(
	parentHash []byte,
	blockNumber uint32,
	stateRoot []byte,
	extrinsicsRoot []byte,
	digest []byte,
	root commitmenttypes.MerkleRoot,
	timestamp time.Time,
) *ConsensusState {
	return &ConsensusState{
		/// The parent hash.
		ParentHash: parentHash,
		/// The block number.
		BlockNumber: blockNumber,
		/// The state trie merkle root
		StateRoot: stateRoot,
		/// The merkle root of the extrinsics.
		ExtrinsicsRoot: extrinsicsRoot,
		/// A chain-specific digest of data useful for light clients or referencing auxiliary data.
		Digest:    digest,
		Root:      root,
		Timestamp: timestamp,
	}
}

// ClientType returns Grandpa
func (ConsensusState) ClientType() string {
	return exported.Grandpa
}

// GetRoot returns the commitment Root for the specific
func (cs ConsensusState) GetRoot() exported.Root {
	// return commitmenttypes.NewMerkleRoot([]byte(SentinelRoot))
	return cs.Root

}

// GetTimestamp returns block time in nanoseconds of the header that created consensus state
func (cs ConsensusState) GetTimestamp() uint64 {
	
	return uint64(cs.Timestamp.UnixNano())

}

// ValidateBasic defines a basic validation for the tendermint consensus state.
// NOTE: ProcessedTimestamp may be zero if this is an initial consensus state passed in by relayer
// as opposed to a consensus state constructed by the chain.
func (cs ConsensusState) ValidateBasic() error {
	// if cs.Root.Empty() {
	// 	return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "root cannot be empty")
	// }
	// if err := tmtypes.ValidateHash(cs.NextValidatorsHash); err != nil {
	// 	return sdkerrors.Wrap(err, "next validators hash is invalid")
	// }
	// if cs.Timestamp.Unix() <= 0 {
	// 	return sdkerrors.Wrap(clienttypes.ErrInvalidConsensus, "timestamp must be a positive Unix time")
	// }
	return nil
}

package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
)

// TODO: impl CheckMisbehaviourAndUpdateState
// CheckMisbehaviourAndUpdateState determines whether or not two conflicting
// headers at the same height would have convinced the light client.
//
// NOTE: consensusState1 is the trusted consensus state that corresponds to the TrustedHeight
// of misbehaviour.Header1
// Similarly, consensusState2 is the trusted consensus state that corresponds
// to misbehaviour.Header2
// Misbehaviour sets frozen height to {0, 1} since it is only used as a boolean value (zero or non-zero).
func (cs ClientState) CheckMisbehaviourAndUpdateState(
	ctx sdk.Context,
	cdc codec.BinaryCodec,
	clientStore sdk.KVStore,
	misbehaviour exported.Misbehaviour,
) (exported.ClientState, error) {
	return &cs, nil
}

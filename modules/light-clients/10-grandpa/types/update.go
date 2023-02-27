package types

import (
	"errors"
	"fmt"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/octopus-network/beefy-go/beefy"
)

// CheckHeaderAndUpdateState checks if the provided header is valid, and if valid it will:
// create the consensus state for the header.Height
// and update the client state if the header height is greater than the latest client state height
// It returns an error if:
// - the client or header provided are not parseable to tendermint types
// - the header is invalid
// - header height is less than or equal to the trusted header height
// - header revision is not equal to trusted header revision
// - header valset commit verification fails
// - header timestamp is past the trusting period in relation to the consensus state
// - header timestamp is less than or equal to the consensus state timestamp
//
// UpdateClient may be used to either create a consensus state for:
// - a future height greater than the latest client state height
// - a past height that was skipped during bisection
// If we are updating to a past height, a consensus state is created for that height to be persisted in client store
// If we are updating to a future height, the consensus state is created and the client state is updated to reflect
// the new latest height
// UpdateClient must only be used to update within a single revision, thus header revision number and trusted height's revision
// number must be the same. To update to a new revision, use a separate upgrade path
// Tendermint client validity checking uses the bisection algorithm described
// in the [Tendermint spec](https://github.com/tendermint/spec/blob/master/spec/consensus/light-client.md).
//
// Misbehaviour Detection:
// UpdateClient will detect implicit misbehaviour by enforcing certain invariants on any new update call and will return a frozen client.
// 1. Any valid update that creates a different consensus state for an already existing height is evidence of misbehaviour and will freeze client.
// 2. Any valid update that breaks time monotonicity with respect to its neighboring consensus states is evidence of misbehaviour and will freeze client.
// Misbehaviour sets frozen height to {0, 1} since it is only used as a boolean value (zero or non-zero).
//
// Pruning:
// UpdateClient will additionally retrieve the earliest consensus state for this clientID and check if it is expired. If it is,
// that consensus state will be pruned from store along with all associated metadata. This will prevent the client store from
// becoming bloated with expired consensus states that can no longer be used for updates and packet verification.
func (cs ClientState) CheckHeaderAndUpdateState(
	ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore,
	header exported.Header,
) (exported.ClientState, exported.ConsensusState, error) {
	fmt.Println("[Grandpa]************Grandpa client CheckHeaderAndUpdateState begin ****************")
	gpHeader, ok := header.(*Header)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader, "expected type %T, got %T", &Header{}, header,
		)
	}

	beefyMMR := gpHeader.BeefyMmr

	// step1:  verify signature
	gsc := beefyMMR.SignedCommitment
	beefyPalyloads := make([]types.PayloadItem, len(gsc.Commitment.Payloads))

	// convert payloads
	for i, v := range gsc.Commitment.Payloads {
		beefyPalyloads[i] = types.PayloadItem{
			ID:   beefy.Bytes2(v.Id),
			Data: v.Data,
		}
	}
	// convert signature
	beefySignatures := make([]beefy.Signature, len(gsc.Signatures))
	for i, v := range gsc.Signatures {
		beefySignatures[i] = beefy.Signature{
			Index:     v.Index,
			Signature: v.Signature,
		}
	}

	// build beefy SignedCommitment
	bsc := beefy.SignedCommitment{
		Commitment: types.Commitment{
			Payload:        beefyPalyloads,
			BlockNumber:    gsc.Commitment.BlockNumber,
			ValidatorSetID: gsc.Commitment.ValidatorSetId,
		},
		Signatures: beefySignatures,
	}
	// checking signatures Threshold
	if beefy.SignatureThreshold(cs.AuthoritySet.Len) > uint32(len(gsc.Signatures)) ||
		beefy.SignatureThreshold(cs.NextAuthoritySet.Len) > uint32(len(gsc.Signatures)) {

		return nil, nil, sdkerrors.Wrap(errors.New(""), ErrInvalidValidatorSet.Error())

	}

	// verify signatures
	switch gsc.Commitment.ValidatorSetId {
	case cs.AuthoritySet.Id:
		err := beefy.VerifySignature(bsc, uint64(cs.AuthoritySet.Len), beefy.Bytes32(cs.AuthoritySet.Root), beefyMMR.SignatureProofs)
		if err != nil {
			return nil, nil, sdkerrors.Wrap(err, ErrInvalidValidatorSet.Error())
		}

	case cs.NextAuthoritySet.Id:
		err := beefy.VerifySignature(bsc, uint64(cs.NextAuthoritySet.Len), beefy.Bytes32(cs.NextAuthoritySet.Root), beefyMMR.SignatureProofs)
		if err != nil {
			return nil, nil, sdkerrors.Wrap(err, ErrInvalidValidatorSet.Error())
		}

	}
	// step2: verify mmr
	// convert mmrleaf
	beefyMMRLeaves := make([]types.MMRLeaf, len(beefyMMR.MmrLeavesAndBatchProof.Leaves))
	for i, v := range beefyMMR.MmrLeavesAndBatchProof.Leaves {
		beefyMMRLeaves[i] = types.MMRLeaf{
			Version: types.MMRLeafVersion(v.Version),
			ParentNumberAndHash: types.ParentNumberAndHash{
				ParentNumber: types.U32(v.ParentNumberAndHash.ParentNumber),
				Hash:         types.NewHash(v.ParentNumberAndHash.ParentHash),
			},
			ParachainHeads: types.NewH256(v.ParachainHeads),
		}
	}
	// convert mmr batch proof
	beefyLeafIndexes := make([]types.U64, len(beefyMMR.MmrLeavesAndBatchProof.MmrBatchProof.LeafIndexes))
	for i, v := range beefyMMR.MmrLeavesAndBatchProof.MmrBatchProof.LeafIndexes {
		beefyLeafIndexes[i] = types.NewU64(v)
	}

	beefyItems := make([]types.H256, len(beefyMMR.MmrLeavesAndBatchProof.MmrBatchProof.Items))
	for i, v := range beefyMMR.MmrLeavesAndBatchProof.MmrBatchProof.Items {
		beefyItems[i] = types.NewH256(v)
	}
	beefyBatchProof := beefy.MMRBatchProof{
		LeafIndex: beefyLeafIndexes,
		LeafCount: types.NewU64(beefyMMR.MmrLeavesAndBatchProof.MmrBatchProof.LeafCount),
		Items:     beefyItems,
	}
	// verify mmr batch proof
	if beefyMMR.SignedCommitment.Commitment.BlockNumber > cs.LatestBeefyHeight {
		result, err := beefy.VerifyMMRBatchProof(beefyPalyloads, beefyMMR.MmrSize,
			beefyMMRLeaves, beefyBatchProof)
		if err != nil || !result {
			return nil, nil, sdkerrors.Wrap(err, "failed to verify mmr proof")
		}
	}

	// step3: verify header
	// headerMap := gpHeader.Message
	switch cs.ChainType {
	case beefy.SOLOCHAIN:

		// convert pb leaf to substrate mmr leaf
		beefySolochainHeaderMap := make(map[uint32]beefy.SolochainHeader)
		headerMap := gpHeader.GetSolochainHeaderMap()
		for num, header := range headerMap.SolochainHeaderMap {
			beefySolochainHeaderMap[num] = beefy.SolochainHeader{
				BlockHeader: header.BlockHeader,
				Timestamp:   beefy.StateProof(header.Timestamp),
			}
		}
		beefy.VerifySolochainHeader(beefyMMRLeaves, beefySolochainHeaderMap)
		//TODO: update consensue state
		
		
	case beefy.PARACHAIN:
		
		// convert pb leaf to substrate mmr leaf
		beefyParachainHeaderMap := make(map[uint32]beefy.ParachainHeader)
		headerMap := gpHeader.GetParachainHeaderMap()

		for num, header := range headerMap.ParachainHeaderMap {

			// convert proofs
			// proofs := make([][]byte,len(header.Proofs))
			// for i,v :=range header.Proofs{
			// 	proofs[i]=v
			// }

			beefyParachainHeaderMap[num] = beefy.ParachainHeader{
				ParaId:      header.ParachainId,
				BlockHeader: header.BlockHeader,
				Proof:       header.Proofs,
				HeaderIndex: header.HeaderIndex,
				HeaderCount: header.HeaderCount,
				Timestamp:   beefy.StateProof(header.Timestamp),
			}

		}
		beefy.VerifyParachainHeader(beefyMMRLeaves, beefyParachainHeaderMap)
		//TODO: update consensue state
		//TODO: save all the consensue state at height,but just return only one
	}

	

	//TODO:finally,build and return new client state ,new consensus state
	// newClientState, newConsensusState := update(ctx, clientStore, &cs, tmHeader)
	newClientState := NewClientState(cs.ChainId, cs.ChainType, cs.BeefyActivationBlock, cs.LatestBeefyHeight,
		cs.MmrRootHash, cs.LatestHeight, cs.FrozenHeight, cs.AuthoritySet, cs.NextAuthoritySet)
	newConsensusState, _ := GetConsensusState(clientStore, cdc, cs.GetLatestHeight())
	// set metadata for this consensus state
	setConsensusMetadata(ctx, clientStore, gpHeader.GetHeight())
	fmt.Println("[Grandpa] new client state")
	fmt.Println(newClientState)
	fmt.Println("[Grandpa] new client consensusState")
	fmt.Println(newConsensusState)
	fmt.Println("[Grandpa]************Grandpa client CheckHeaderAndUpdateState end ****************")
	return newClientState, newConsensusState, nil
}

// checkTrustedHeader checks that consensus state matches trusted fields of Header
func checkTrustedHeader(header *Header, consState *ConsensusState) error {
	// tmTrustedValidators, err := tmtypes.ValidatorSetFromProto(header.TrustedValidators)
	// if err != nil {
	// 	return sdkerrors.Wrap(err, "trusted validator set in not tendermint validator set type")
	// }

	// // assert that trustedVals is NextValidators of last trusted header
	// // to do this, we check that trustedVals.Hash() == consState.NextValidatorsHash
	// tvalHash := tmTrustedValidators.Hash()
	// if !bytes.Equal(consState.NextValidatorsHash, tvalHash) {
	// 	return sdkerrors.Wrapf(
	// 		ErrInvalidValidatorSet,
	// 		"trusted validators %s, does not hash to latest trusted validators. Expected: %X, got: %X",
	// 		header.TrustedValidators, consState.NextValidatorsHash, tvalHash,
	// 	)
	// }
	return nil
}

// checkValidity checks if the Grandpa header is valid.
// CONTRACT: consState.Height == header.TrustedHeight
func checkValidity(
	clientState *ClientState, consState *ConsensusState,
	header *Header, currentTimestamp time.Time,
) error {
	// if err := checkTrustedHeader(header, consState); err != nil {
	// 	return err
	// }

	// // UpdateClient only accepts updates with a header at the same revision
	// // as the trusted consensus state
	// if header.GetHeight().GetRevisionNumber() != header.TrustedHeight.RevisionNumber {
	// 	return sdkerrors.Wrapf(
	// 		ErrInvalidHeaderHeight,
	// 		"header height revision %d does not match trusted header revision %d",
	// 		header.GetHeight().GetRevisionNumber(), header.TrustedHeight.RevisionNumber,
	// 	)
	// }

	// tmTrustedValidators, err := tmtypes.ValidatorSetFromProto(header.TrustedValidators)
	// if err != nil {
	// 	return sdkerrors.Wrap(err, "trusted validator set in not tendermint validator set type")
	// }

	// tmSignedHeader, err := tmtypes.SignedHeaderFromProto(header.SignedHeader)
	// if err != nil {
	// 	return sdkerrors.Wrap(err, "signed header in not tendermint signed header type")
	// }

	// tmValidatorSet, err := tmtypes.ValidatorSetFromProto(header.ValidatorSet)
	// if err != nil {
	// 	return sdkerrors.Wrap(err, "validator set in not tendermint validator set type")
	// }

	// // assert header height is newer than consensus state
	// if header.GetHeight().LTE(header.TrustedHeight) {
	// 	return sdkerrors.Wrapf(
	// 		clienttypes.ErrInvalidHeader,
	// 		"header height ≤ consensus state height (%s ≤ %s)", header.GetHeight(), header.TrustedHeight,
	// 	)
	// }

	// chainID := clientState.GetChainID()
	// // If chainID is in revision format, then set revision number of chainID with the revision number
	// // of the header we are verifying
	// // This is useful if the update is at a previous revision rather than an update to the latest revision
	// // of the client.
	// // The chainID must be set correctly for the previous revision before attempting verification.
	// // Updates for previous revisions are not supported if the chainID is not in revision format.
	// if clienttypes.IsRevisionFormat(chainID) {
	// 	chainID, _ = clienttypes.SetRevisionNumber(chainID, header.GetHeight().GetRevisionNumber())
	// }

	// // Construct a trusted header using the fields in consensus state
	// // Only Height, Time, and NextValidatorsHash are necessary for verification
	// trustedHeader := tmtypes.Header{
	// 	ChainID: chainID,
	// 	Height:  int64(header.RevisionHeight),
	// }
	// signedHeader := tmtypes.SignedHeader{
	// 	Header: &trustedHeader,
	// }

	// // Verify next header with the passed-in trustedVals
	// // - asserts trusting period not passed
	// // - assert header timestamp is not past the trusting period
	// // - assert header timestamp is past latest stored consensus state timestamp
	// // - assert that a TrustLevel proportion of TrustedValidators signed new Commit
	// err = light.Verify(
	// 	&signedHeader,
	// 	tmTrustedValidators, tmSignedHeader, tmValidatorSet,
	// 	clientState.TrustingPeriod, currentTimestamp, clientState.MaxClockDrift, clientState.TrustLevel.ToTendermint(),
	// )
	// if err != nil {
	// 	return sdkerrors.Wrap(err, "failed to verify header")
	// }
	return nil
}

// update the consensus state from a new header and set processed time metadata
func update(ctx sdk.Context, clientStore sdk.KVStore, clientState *ClientState, header *Header) (*ClientState, *ConsensusState) {
	// height := header.GetHeight().(clienttypes.Height)
	// if height.GT(clientState.LatestHeight) {
	// 	clientState.LatestHeight = height

	// }

	// clientState.LatestHeight = height
	// clientState.BlockNumber = header.BlockHeader.BlockNumber
	// clientState.BlockHeader = header.BlockHeader

	consensusState := &ConsensusState{
		/// The parent hash.

		Root: []byte{},
		// Timestamp: time.Now(),
		Timestamp: time.Unix(0, 0),
	}

	// set metadata for this consensus state
	setConsensusMetadata(ctx, clientStore, header.GetHeight())

	fmt.Printf("[Grandpa]new header is %s \n", header.String())
	fmt.Printf("[Grandpa] new client state latestheight is %d \n", clientState.LatestHeight)

	return clientState, consensusState
}


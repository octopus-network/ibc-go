package types

import (
	"errors"
	"time"

	// "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	gsrpctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	gsrpccodec "github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/octopus-network/beefy-go/beefy"
)

// step1: verify signature
// step2: verify mmr
// step3: verfiy header
// step4: update client state and consensuse state
func (cs ClientState) CheckHeaderAndUpdateState(
	ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore,
	header exported.Header,
) (exported.ClientState, exported.ConsensusState, error) {
	ctx.Logger().Debug("LightClient:", "10-Grandpa", "method:", "CheckHeaderAndUpdateState")

	pbHeader, ok := header.(*Header)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader, "expected type %T, got %T", &Header{}, header,
		)
	}

	beefyMMR := pbHeader.BeefyMmr

	// step1:  verify signature
	// convert signedcommitment
	bsc := Convert2BeefySC(beefyMMR.SignedCommitment)
	err := cs.VerifySignatures(bsc, beefyMMR.SignatureProofs)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "failed to verify signatures")
	}

	// step2: verify mmr
	// convert mmrleaf
	beefyMMRLeaves := Convert2BeefyMMRLeaves(beefyMMR.MmrLeavesAndBatchProof.Leaves)
	// convert mmr proof
	beefyBatchProof := Convert2MMRBatchProof(beefyMMR.MmrLeavesAndBatchProof)
	// verify mmr
	err = cs.VerifyMMR(bsc, beefyMMR.MmrSize, beefyMMRLeaves, beefyBatchProof)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "failed to verify mmr")
	}

	// step3: verify header
	err = cs.VerifyHeader(*pbHeader, beefyMMRLeaves)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "failed to verify header")
	}

	// step4: update state
	// update client state and build new client state
	newClientState, err := cs.UpdateClientState(ctx, beefyMMR.SignedCommitment.Commitment, beefyMMR.MmrLeavesAndBatchProof.Leaves)
	if err != nil {
		return nil, nil, err
	}
	// update consensue state and build latest new consensue state
	newConsensusState, err := cs.UpdateConsensusStates(ctx, cdc, clientStore, pbHeader)
	if err != nil {
		return nil, nil, err
	}
	return newClientState, newConsensusState, nil
}

// verify signatures
func (cs ClientState) VerifySignatures(bsc beefy.SignedCommitment, SignatureProofs [][]byte) error {

	// checking signatures Threshold
	if beefy.SignatureThreshold(cs.AuthoritySet.Len) > uint32(len(bsc.Signatures)) ||
		beefy.SignatureThreshold(cs.NextAuthoritySet.Len) > uint32(len(bsc.Signatures)) {

		return sdkerrors.Wrap(errors.New("verify signature error "), ErrInvalidValidatorSet.Error())

	}
	// verify signatures
	switch bsc.Commitment.ValidatorSetID {
	case cs.AuthoritySet.Id:
		err := beefy.VerifySignature(bsc, uint64(cs.AuthoritySet.Len), beefy.Bytes32(cs.AuthoritySet.Root), SignatureProofs)
		if err != nil {
			return sdkerrors.Wrap(err, ErrInvalidValidatorSet.Error())
		}

	case cs.NextAuthoritySet.Id:
		err := beefy.VerifySignature(bsc, uint64(cs.NextAuthoritySet.Len), beefy.Bytes32(cs.NextAuthoritySet.Root), SignatureProofs)
		if err != nil {
			return sdkerrors.Wrap(err, ErrInvalidValidatorSet.Error())
		}

	}

	return nil
}

// verify batch mmr proof
func (cs ClientState) VerifyMMR(bsc beefy.SignedCommitment, mmrSize uint64,
	beefyMMRLeaves []types.MMRLeaf, mmrBatchProof beefy.MMRBatchProof) error {
	// check mmr height
	if bsc.Commitment.BlockNumber <= cs.LatestBeefyHeight {
		return sdkerrors.Wrap(errors.New("Commitment.BlockNumber < cs.LatestBeefyHeight"), "")
	}

	//verify mmr proof
	result, err := beefy.VerifyMMRBatchProof(bsc.Commitment.Payload, mmrSize, beefyMMRLeaves, mmrBatchProof)
	if err != nil || !result {
		return sdkerrors.Wrap(errors.New("failed to verify mmr proof"), "")
	}
	return nil
}

//  verify solochain header or parachain header
func (cs ClientState) VerifyHeader(gpHeader Header, beefyMMRLeaves []types.MMRLeaf,
) error {

	switch cs.ChainType {
	case beefy.CHAINTYPE_SOLOCHAIN:
		Logger.Debug("verify solochain header")

		headerMap := gpHeader.GetSolochainHeaderMap()
		// convert pb solochain header to beefy solochain header
		beefySolochainHeaderMap := make(map[uint32]beefy.SolochainHeader)
		for num, header := range headerMap.SolochainHeaderMap {
			beefySolochainHeaderMap[num] = beefy.SolochainHeader{
				BlockHeader: header.BlockHeader,
				Timestamp:   beefy.StateProof(header.Timestamp),
			}
		}
		err := beefy.VerifySolochainHeader(beefyMMRLeaves, beefySolochainHeaderMap)
		if err != nil {
			return err
		}

	case beefy.CHAINTYPE_PARACHAIN:
		Logger.Debug("verify parachain header")

		headerMap := gpHeader.GetParachainHeaderMap()
		// convert pb parachain header to beefy parachain header
		beefyParachainHeaderMap := make(map[uint32]beefy.ParachainHeader)
		for num, header := range headerMap.ParachainHeaderMap {
			beefyParachainHeaderMap[num] = beefy.ParachainHeader{
				ParaId:      header.ParachainId,
				BlockHeader: header.BlockHeader,
				Proof:       header.Proofs,
				HeaderIndex: header.HeaderIndex,
				HeaderCount: header.HeaderCount,
				Timestamp:   beefy.StateProof(header.Timestamp),
			}
		}
		err := beefy.VerifyParachainHeader(beefyMMRLeaves, beefyParachainHeaderMap)
		if err != nil {
			return err
		}
	}
	return nil
}

// update client state
func (cs ClientState) UpdateClientState(ctx sdk.Context, commitment Commitment, mmrLeaves []MMRLeaf) (*ClientState, error) {
	ctx.Logger().Debug("update client state")
	// var newClientState *ClientState

	latestBeefyHeight := commitment.BlockNumber
	latestChainHeight := latestBeefyHeight
	mmrRoot := commitment.Payloads[0].Data

	// find latest next authority set from mmrleaves
	var latestNextAuthoritySet *BeefyAuthoritySet
	var latestAuthoritySetId uint64
	for _, leaf := range mmrLeaves {
		if latestAuthoritySetId < leaf.BeefyNextAuthoritySet.Id {
			latestAuthoritySetId = leaf.BeefyNextAuthoritySet.Id
			latestNextAuthoritySet = &BeefyAuthoritySet{
				Id:   leaf.BeefyNextAuthoritySet.Id,
				Len:  leaf.BeefyNextAuthoritySet.Len,
				Root: leaf.BeefyNextAuthoritySet.Root,
			}
		}

	}

	newClientState := NewClientState(cs.ChainType,cs.ChainId,cs.ParachainId, cs.BeefyActivationBlock,
		latestBeefyHeight, mmrRoot, latestChainHeight, cs.FrozenHeight,
		*latestNextAuthoritySet, cs.NextAuthoritySet)

	return newClientState, nil

}

//TODO: save all the consensue state at height,but just return only one
func (cs ClientState) UpdateConsensusStates(ctx sdk.Context, cdc codec.BinaryCodec, clientStore sdk.KVStore, header *Header) (*ConsensusState, error) {

	var newConsensueState *ConsensusState
	var latestChainHeight uint32
	var latestBlockHeader *gsrpctypes.Header
	var latestTimestamp uint64
	switch cs.ChainType {
	case beefy.CHAINTYPE_SOLOCHAIN:
		solochainHeaderMap := header.GetSolochainHeaderMap().SolochainHeaderMap

		for _, header := range solochainHeaderMap {
			var decodeHeader gsrpctypes.Header
			err := gsrpccodec.Decode(header.BlockHeader, &decodeHeader)
			if err != nil {
				return nil, sdkerrors.Wrapf(err, "decode header error")
			}
			var decodeTimestamp gsrpctypes.U64
			err = gsrpccodec.Decode(header.Timestamp.Value, &decodeTimestamp)
			if err != nil {
				return nil, sdkerrors.Wrapf(err, "decode timestamp error")
			}
			err = updateConsensuestate(clientStore, cdc, decodeHeader, uint64(decodeTimestamp))
			if err != nil {
				return nil, sdkerrors.Wrapf(clienttypes.ErrFailedClientConsensusStateVerification, "update consensus state failed")
			}
			if latestChainHeight < uint32(decodeHeader.Number) {
				latestChainHeight = uint32(decodeHeader.Number)
				latestBlockHeader = &decodeHeader
				latestTimestamp = uint64(decodeTimestamp)
			}
		}

		//find latest header and build new consensue state
		newConsensueState = NewConsensusState(latestBlockHeader.StateRoot[:], time.UnixMilli(int64(latestTimestamp)))

		// set metadata for this consensus
		latestHeigh := clienttypes.Height{
			RevisionNumber: 0,
			RevisionHeight: uint64(latestChainHeight),
		}
		setConsensusMetadata(ctx, clientStore, latestHeigh)

	case beefy.CHAINTYPE_PARACHAIN:
		parachainHeaderMap := header.GetParachainHeaderMap().ParachainHeaderMap
		// var latestHeight uint32
		// var latestHeader *gsrpctypes.Header
		// var latestTimestamp uint64
		for _, header := range parachainHeaderMap {
			var decodeHeader gsrpctypes.Header
			err := gsrpccodec.Decode(header.BlockHeader, &decodeHeader)
			if err != nil {
				return nil, sdkerrors.Wrapf(err, "decode header error")
			}
			var decodeTimestamp gsrpctypes.U64
			err = gsrpccodec.Decode(header.Timestamp.Value, &decodeTimestamp)
			if err != nil {
				return nil, sdkerrors.Wrapf(err, "decode timestamp error")
			}
			err = updateConsensuestate(clientStore, cdc, decodeHeader, uint64(decodeTimestamp))
			if err != nil {
				return nil, sdkerrors.Wrapf(clienttypes.ErrFailedClientConsensusStateVerification, "update consensus state failed")
			}

			if latestChainHeight < uint32(decodeHeader.Number) {
				latestChainHeight = uint32(decodeHeader.Number)
				latestBlockHeader = &decodeHeader
				latestTimestamp = uint64(decodeTimestamp)
			}

		}
		//find latest header and build new consensue state
		newConsensueState = NewConsensusState(latestBlockHeader.StateRoot[:], time.UnixMilli(int64(latestTimestamp)))

		// set metadata for this consensus
		latestHeigh := clienttypes.Height{
			RevisionNumber: 0,
			RevisionHeight: uint64(latestChainHeight),
		}
		setConsensusMetadata(ctx, clientStore, latestHeigh)
	}

	// TODO: consider to pruning!
	//
	return newConsensueState, nil

}

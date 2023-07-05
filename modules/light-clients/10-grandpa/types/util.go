package types

import (
	gsrpctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/octopus-network/beefy-go/beefy"
)

// Type conversion,from beefy mmr to pb(protocol buffer) mmr
func ToPBBeefyMMR(bsc beefy.SignedCommitment, mmrBatchProof beefy.MmrProofsResp, authorityProof [][]byte) BeefyMMR {

	// bsc := beefy.ConvertCommitment(sc)
	pbPalyloads := make([]PayloadItem, len(bsc.Commitment.Payload))
	for i, v := range bsc.Commitment.Payload {
		pbPalyloads[i] = PayloadItem{
			Id:   v.ID[:],
			Data: v.Data,
		}

	}

	pbCommitment := Commitment{
		Payloads:       pbPalyloads,
		BlockNumber:    bsc.Commitment.BlockNumber,
		ValidatorSetId: bsc.Commitment.ValidatorSetID,
	}

	pb := make([]Signature, len(bsc.Signatures))
	for i, v := range bsc.Signatures {
		pb[i] = Signature(v)
	}

	pbsc := SignedCommitment{
		Commitment: pbCommitment,
		Signatures: pb,
	}

	pbMmrLevavesAndProof := ToPbMmrProof(mmrBatchProof)

	// leafIndex := beefy.ConvertBlockNumberToMmrLeafIndex(uint32(beefy.BEEFY_ACTIVATION_BLOCK), bsc.Commitment.BlockNumber)
	// mmrSize := mmr.LeafIndexToMMRSize(uint64(leafIndex))
	// build pbBeefyMMR
	pbBeefyMMR := BeefyMMR{
		SignedCommitment:       pbsc,
		SignatureProofs:        authorityProof,
		MmrLeavesAndBatchProof: pbMmrLevavesAndProof,
	}
	return pbBeefyMMR
}

// Type conversion,from beefy mmr proof to pb(protocol buffer) mmr proof
func ToPbMmrProof(mmrBatchProof beefy.MmrProofsResp) MMRLeavesAndBatchProof {
	// convert mmrleaf
	var pbMMRLeaves []MMRLeaf

	leafNum := len(mmrBatchProof.Leaves)
	for i := 0; i < leafNum; i++ {
		leaf := mmrBatchProof.Leaves[i]
		parentNumAndHash := ParentNumberAndHash{
			ParentNumber: uint32(leaf.ParentNumberAndHash.ParentNumber),
			ParentHash:   []byte(leaf.ParentNumberAndHash.Hash[:]),
		}
		nextAuthoritySet := BeefyAuthoritySet{
			Id:   uint64(leaf.BeefyNextAuthoritySet.ID),
			Len:  uint32(leaf.BeefyNextAuthoritySet.Len),
			Root: []byte(leaf.BeefyNextAuthoritySet.Root[:]),
		}
		parachainHeads := []byte(leaf.ParachainHeads[:])
		gLeaf := MMRLeaf{
			Version:               uint32(leaf.Version),
			ParentNumberAndHash:   parentNumAndHash,
			BeefyNextAuthoritySet: nextAuthoritySet,
			ParachainHeads:        parachainHeads,
		}
		// Logger.Info("gLeaf: ", gLeaf)
		pbMMRLeaves = append(pbMMRLeaves, gLeaf)
	}

	// convert mmr batch proof
	pbLeafIndexes := make([]uint64, len(mmrBatchProof.Proof.LeafIndexes))
	for i, v := range mmrBatchProof.Proof.LeafIndexes {
		pbLeafIndexes[i] = uint64(v)
	}

	pbProofItems := [][]byte{}
	itemNum := len(mmrBatchProof.Proof.Items)
	for i := 0; i < itemNum; i++ {
		item := mmrBatchProof.Proof.Items[i][:]
		pbProofItems = append(pbProofItems, item)

	}

	pbBatchProof := MMRBatchProof{
		LeafIndexes: pbLeafIndexes,
		LeafCount:   uint64(mmrBatchProof.Proof.LeafCount),
		Items:       pbProofItems,
	}

	pbMmrLevavesAndProof := MMRLeavesAndBatchProof{
		Leaves:        pbMMRLeaves,
		MmrBatchProof: pbBatchProof,
	}
	return pbMmrLevavesAndProof
}

// Type conversion,from  pb(protocol buffer) SignedCommitment to beefy SignedCommitment
func ToBeefySC(pbsc SignedCommitment) beefy.SignedCommitment {
	beefyPalyloads := make([]gsrpctypes.PayloadItem, len(pbsc.Commitment.Payloads))
	// // step1:  verify signature
	for i, v := range pbsc.Commitment.Payloads {
		beefyPalyloads[i] = gsrpctypes.PayloadItem{
			ID:   beefy.Bytes2(v.Id),
			Data: v.Data,
		}
	}
	// convert signature
	beefySignatures := make([]beefy.Signature, len(pbsc.Signatures))
	for i, v := range pbsc.Signatures {
		beefySignatures[i] = beefy.Signature{
			Index:     v.Index,
			Signature: v.Signature,
		}
	}
	// build beefy SignedCommitment
	bsc := beefy.SignedCommitment{
		Commitment: gsrpctypes.Commitment{
			Payload:        beefyPalyloads,
			BlockNumber:    pbsc.Commitment.BlockNumber,
			ValidatorSetID: pbsc.Commitment.ValidatorSetId,
		},
		Signatures: beefySignatures,
	}

	return bsc
}

// Type conversion,from  pb(protocol buffer) MMRLeaf to beefy MMRLeaf
func ToBeefyMMRLeaves(pbMMRLeaves []MMRLeaf) []gsrpctypes.MMRLeaf {

	beefyMMRLeaves := make([]gsrpctypes.MMRLeaf, len(pbMMRLeaves))
	for i, v := range pbMMRLeaves {
		beefyMMRLeaves[i] = gsrpctypes.MMRLeaf{
			Version: gsrpctypes.MMRLeafVersion(v.Version),
			ParentNumberAndHash: gsrpctypes.ParentNumberAndHash{
				ParentNumber: gsrpctypes.U32(v.ParentNumberAndHash.ParentNumber),
				Hash:         gsrpctypes.NewHash(v.ParentNumberAndHash.ParentHash),
			},
			BeefyNextAuthoritySet: gsrpctypes.BeefyNextAuthoritySet{
				ID:   gsrpctypes.U64(v.BeefyNextAuthoritySet.Id),
				Len:  gsrpctypes.U32(v.BeefyNextAuthoritySet.Len),
				Root: gsrpctypes.NewH256(v.BeefyNextAuthoritySet.Root),
			},
			ParachainHeads: gsrpctypes.NewH256(v.ParachainHeads),
		}
	}

	return beefyMMRLeaves
}

func ToMMRBatchProof(mmrBatchProof MMRBatchProof) beefy.MMRBatchProof {
	pbLeafIndexes := mmrBatchProof.LeafIndexes
	leafIndexes := make([]gsrpctypes.U64, len(pbLeafIndexes))
	for i, v := range pbLeafIndexes {
		leafIndexes[i] = gsrpctypes.NewU64(v)
	}

	pbItems := mmrBatchProof.Items
	items := make([]gsrpctypes.H256, len(pbItems))
	for i, v := range pbItems {
		items[i] = gsrpctypes.NewH256(v)
	}

	beefyMmrBatchProof := beefy.MMRBatchProof{
		LeafIndexes: leafIndexes,
		LeafCount:   gsrpctypes.NewU64(mmrBatchProof.LeafCount),
		Items:       items,
	}

	return beefyMmrBatchProof

}

// Type conversion,from  beefy SubchainHeader to pb SubchainHeader
func ToPBSubchainHeaders(beefySubchainHeaders []beefy.SubchainHeader, mmrBatchProof beefy.MmrProofsResp) Header_SubchainHeaders {

	// headerMap := make(map[uint32]SubchainHeader)
	var headers []SubchainHeader
	for _, header := range beefySubchainHeaders {
		subHeader := SubchainHeader{
			ChainId:     header.ChainId,
			BlockNumber: header.BlockNumber,
			BlockHeader: header.BlockHeader,
			Timestamp:   StateProof(header.Timestamp),
		}
		headers = append(headers, subHeader)

	}

	pbMmrLevavesAndProof := ToPbMmrProof(mmrBatchProof)
	pbSubchainHeaders := SubchainHeaders{
		SubchainHeaders:        headers,
		MmrLeavesAndBatchProof: pbMmrLevavesAndProof,
	}

	header_subchainHeaders := Header_SubchainHeaders{
		SubchainHeaders: &pbSubchainHeaders,
	}
	return header_subchainHeaders

}

// Type conversion,from  beefy ParachainHeader to pb ParachainHeader
func ToPBParachainHeaders(beefyParachainHeaders []beefy.ParachainHeader, mmrBatchProof beefy.MmrProofsResp) Header_ParachainHeaders {

	var headers []ParachainHeader
	for _, header := range beefyParachainHeaders {
		parachainHeader := ParachainHeader{
			ChainId:            header.ChainId,
			ParachainId:        header.ParaId,
			RelayerChainNumber: header.RelayerChainNumber,
			BlockHeader:        header.BlockHeader,
			Proofs:             header.Proof,
			HeaderIndex:        header.HeaderIndex,
			HeaderCount:        header.HeaderCount,
			Timestamp:          StateProof(header.Timestamp),
		}
		headers = append(headers, parachainHeader)
	}
	pbMmrLevavesAndProof := ToPbMmrProof(mmrBatchProof)
	pbParachainHeaders := ParachainHeaders{
		ParachainHeaders:       headers,
		MmrLeavesAndBatchProof: pbMmrLevavesAndProof,
	}

	header_parachainHeaders := Header_ParachainHeaders{
		ParachainHeaders: &pbParachainHeaders,
	}
	return header_parachainHeaders

}

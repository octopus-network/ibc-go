package types_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/ComposableFi/go-merkle-trees/mmr"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/octopus-network/beefy-go/beefy"
	"github.com/stretchr/testify/suite"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	gsrpctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	gsrpccodec "github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	ibctmtypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"
	ibcgptypes "github.com/cosmos/ibc-go/v3/modules/light-clients/10-grandpa/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	ibctestingmock "github.com/cosmos/ibc-go/v3/testing/mock"
	"github.com/cosmos/ibc-go/v3/testing/simapp"
	"github.com/stretchr/testify/require"
)

const (
	chainID                        = "sub"
	chainIDRevision0               = "sub-revision-0"
	chainIDRevision1               = "sub-revision-1"
	clientID                       = "submainnet"
	trustingPeriod   time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod        time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift    time.Duration = time.Second * 10
)

var (
	height          = clienttypes.NewHeight(0, 4)
	newClientHeight = clienttypes.NewHeight(1, 1)
	upgradePath     = []string{"upgrade", "upgradedIBCState"}
)

type GrandpaTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator

	// testing chains used for convenience and readability
	chainA *ibctesting.TestChain
	chainB *ibctesting.TestChain

	// TODO: deprecate usage in favor of testing package
	ctx          sdk.Context
	cdc          codec.Codec
	privVal      tmtypes.PrivValidator
	valSet       *tmtypes.ValidatorSet
	valsHash     tmbytes.HexBytes
	header       *ibctmtypes.Header
	now          time.Time
	headerTime   time.Time
	clientTime   time.Time
	clientStates []*ibcgptypes.ClientState
}

func (suite *GrandpaTestSuite) initClientStates() {
	cs1 := ibcgptypes.ClientState{
		ChainId:              0,
		ChainType:            beefy.CHAINTYPE_SOLOCHAIN,
		BeefyActivationBlock: beefy.BEEFY_ACTIVATION_BLOCK,
		LatestBeefyHeight:    108,
		MmrRootHash:          []uint8{0x7b, 0xd6, 0x16, 0x8a, 0x24, 0xd1, 0xc1, 0xcf, 0x65, 0xca, 0x6e, 0x6e, 0xfe, 0x71, 0xab, 0x16, 0x95, 0xf9, 0x90, 0xaf, 0x1d, 0xaa, 0x11, 0x92, 0x4, 0xfa, 0xe9, 0x2d, 0xfb, 0x96, 0xfd, 0xa4},
	}
	suite.clientStates = append(suite.clientStates, &cs1)

}
func (suite *GrandpaTestSuite) SetupTest() {
	suite.coordinator = ibctesting.NewCoordinator(suite.T(), 2)
	suite.chainA = suite.coordinator.GetChain(ibctesting.GetChainID(1))
	suite.chainB = suite.coordinator.GetChain(ibctesting.GetChainID(2))
	// commit some blocks so that QueryProof returns valid proof (cannot return valid query if height <= 1)
	suite.coordinator.CommitNBlocks(suite.chainA, 2)
	suite.coordinator.CommitNBlocks(suite.chainB, 2)

	// TODO: deprecate usage in favor of testing package
	checkTx := false
	app := simapp.Setup(checkTx)

	suite.cdc = app.AppCodec()

	// now is the time of the current chain, must be after the updating header
	// mocks ctx.BlockTime()
	suite.now = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	suite.clientTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	// Header time is intended to be time for any new header used for updates
	suite.headerTime = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)

	suite.privVal = ibctestingmock.NewPV()

	pubKey, err := suite.privVal.GetPubKey()
	suite.Require().NoError(err)

	heightMinus1 := clienttypes.NewHeight(0, height.RevisionHeight-1)

	val := tmtypes.NewValidator(pubKey, 10)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{val})
	suite.valsHash = suite.valSet.Hash()
	//TODO: change to grandpa
	suite.header = suite.chainA.CreateTMClientHeader(chainID, int64(height.RevisionHeight), heightMinus1, suite.now, suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})
	suite.ctx = app.BaseApp.NewContext(checkTx, tmproto.Header{Height: 1, Time: suite.now})
	suite.initClientStates()
}

func getSuiteSigners(suite *GrandpaTestSuite) []tmtypes.PrivValidator {
	return []tmtypes.PrivValidator{suite.privVal}
}

func getBothSigners(suite *GrandpaTestSuite, altVal *tmtypes.Validator, altPrivVal tmtypes.PrivValidator) (*tmtypes.ValidatorSet, []tmtypes.PrivValidator) {
	// Create bothValSet with both suite validator and altVal. Would be valid update
	bothValSet := tmtypes.NewValidatorSet(append(suite.valSet.Validators, altVal))
	// Create signer array and ensure it is in same order as bothValSet
	_, suiteVal := suite.valSet.GetByIndex(0)
	bothSigners := ibctesting.CreateSortedSignerArray(altPrivVal, suite.privVal, altVal, suiteVal)
	return bothValSet, bothSigners
}

func TestGrandpaTestSuite(t *testing.T) {
	suite.Run(t, new(GrandpaTestSuite))
}

func TestSolochainLocalNet(t *testing.T) {
	// t.Skip("todo: get and build data from substrate chain and invoke the grandpa lc to verify data")
	localSolochainEndpoint, err := gsrpc.NewSubstrateAPI(beefy.LOCAL_RELAY_ENDPPOIT)
	if err != nil {
		t.Logf("Connecting err: %v", err)
	}
	ch := make(chan interface{})
	sub, err := localSolochainEndpoint.Client.Subscribe(
		context.Background(),
		"beefy",
		"subscribeJustifications",
		"unsubscribeJustifications",
		"justifications",
		ch)
	require.NoError(t, err)
	t.Logf("subscribed to %s\n", beefy.LOCAL_RELAY_ENDPPOIT)

	defer sub.Unsubscribe()

	timeout := time.After(24 * time.Hour)
	received := 0
	// var preBlockNumber uint32
	// var preBloackHash gsrpctypes.Hash
	var clientState *ibcgptypes.ClientState
	consensusStateKVStore := make(map[uint32]ibcgptypes.ConsensusState)
	for {
		select {
		case msg := <-ch:
			t.Logf("encoded msg: %s", msg)

			s := &beefy.VersionedFinalityProof{}
			err := gsrpccodec.DecodeFromHex(msg.(string), s)
			if err != nil {
				panic(err)
			}

			t.Logf("decoded msg: %+v\n", s)
			t.Logf("decoded msg: %#v\n", s)
			latestSignedCommitmentBlockNumber := s.SignedCommitment.Commitment.BlockNumber
			// t.Logf("blockNumber: %d\n", latestBlockNumber)
			latestSignedCommitmentBlockHash, err := localSolochainEndpoint.RPC.Chain.GetBlockHash(uint64(latestSignedCommitmentBlockNumber))
			require.NoError(t, err)
			t.Logf("latestSignedCommitmentBlockNumber: %d latestSignedCommitmentBlockHash: %#x",
				latestSignedCommitmentBlockNumber, latestSignedCommitmentBlockHash)

			// build and init grandpa lc client state
			if clientState == nil {
				//  init client state
				t.Log("clientState == nil, need to init !")
				authoritySet, err := beefy.GetBeefyAuthoritySet(latestSignedCommitmentBlockHash, localSolochainEndpoint, "BeefyAuthorities")
				require.NoError(t, err)
				t.Logf("current authority set: %+v", authoritySet)
				nextAuthoritySet, err := beefy.GetBeefyAuthoritySet(latestSignedCommitmentBlockHash, localSolochainEndpoint, "BeefyNextAuthorities")
				require.NoError(t, err)
				t.Logf("next authority set: %+v", nextAuthoritySet)
				// get the parachain or solochain latest height
				// the height must be inclueded into relay chain, if not exist ,the height is zero
				fromBlockNumber := latestSignedCommitmentBlockNumber - 7 // for test
				t.Logf("fromBlockNumber: %d toBlockNumber: %d", fromBlockNumber, latestSignedCommitmentBlockNumber)

				fromBlockHash, err := localSolochainEndpoint.RPC.Chain.GetBlockHash(uint64(fromBlockNumber))
				require.NoError(t, err)
				t.Logf("fromBlockHash: %#x ", fromBlockHash)
				t.Logf("toBlockHash: %#x", latestSignedCommitmentBlockHash)

				clientState = &ibcgptypes.ClientState{
					ChainId:              beefy.LOCAL_SOLOCHAIN_ID,
					ChainType:            beefy.CHAINTYPE_SOLOCHAIN,
					BeefyActivationBlock: beefy.BEEFY_ACTIVATION_BLOCK,
					LatestBeefyHeight:    latestSignedCommitmentBlockNumber,
					MmrRootHash:          s.SignedCommitment.Commitment.Payload[0].Data,
					LatestHeight:         latestSignedCommitmentBlockNumber,
					FrozenHeight:         0,
					AuthoritySet: ibcgptypes.BeefyAuthoritySet{
						Id:   uint64(authoritySet.ID),
						Len:  uint32(authoritySet.Len),
						Root: authoritySet.Root[:],
					},
					NextAuthoritySet: ibcgptypes.BeefyAuthoritySet{
						Id:   uint64(nextAuthoritySet.ID),
						Len:  uint32(nextAuthoritySet.Len),
						Root: nextAuthoritySet.Root[:],
					},
				}
				t.Logf("init client state: %+v", &clientState)
				t.Logf("init client state: %+v", *clientState)
				// test pb marshal and unmarshal
				marshalCS, err := clientState.Marshal()
				require.NoError(t, err)
				t.Logf("marshal client state: %+v", marshalCS)
				// unmarshal
				// err = clientState.Unmarshal(marshalCS)
				var unmarshalCS ibcgptypes.ClientState
				err = unmarshalCS.Unmarshal(marshalCS)
				require.NoError(t, err)
				t.Logf("unmarshal client state: %+v", unmarshalCS)
				continue
			}

			// step1,build authourities proof for current beefy signatures
			authorities, err := beefy.GetBeefyAuthorities(latestSignedCommitmentBlockHash, localSolochainEndpoint, "Authorities")
			require.NoError(t, err)
			bsc := beefy.ConvertCommitment(s.SignedCommitment)
			var authorityIdxes []uint64
			for _, v := range bsc.Signatures {
				idx := v.Index
				authorityIdxes = append(authorityIdxes, uint64(idx))
			}
			_, authorityProof, err := beefy.BuildAuthorityProof(authorities, authorityIdxes)
			require.NoError(t, err)

			// step2,build beefy mmr
			targetHeights := []uint64{uint64(latestSignedCommitmentBlockNumber - 1)}
			// build mmr proofs for leaves containing target paraId
			mmrBatchProof, err := beefy.BuildMMRBatchProof(localSolochainEndpoint, &latestSignedCommitmentBlockHash, targetHeights)
			require.NoError(t, err)

			leafIndex := beefy.ConvertBlockNumberToMmrLeafIndex(uint32(beefy.BEEFY_ACTIVATION_BLOCK), latestSignedCommitmentBlockNumber)
			mmrSize := mmr.LeafIndexToMMRSize(uint64(leafIndex))

			result, err := beefy.VerifyMMRBatchProof(s.SignedCommitment.Commitment.Payload,
				mmrSize, mmrBatchProof.Leaves, mmrBatchProof.Proof)
			require.NoError(t, err)
			require.True(t, result)

			// convert payloads
			gPalyloads := make([]ibcgptypes.PayloadItem, len(bsc.Commitment.Payload))
			for i, v := range bsc.Commitment.Payload {
				gPalyloads[i] = ibcgptypes.PayloadItem{
					Id:   v.ID[:],
					Data: v.Data,
				}

			}
			t.Logf("bsc.Commitment.Payload: %+v", bsc.Commitment.Payload)
			t.Logf("gpPalyloads: %+v", gPalyloads)

			gCommitment := ibcgptypes.Commitment{
				Payloads:       gPalyloads,
				BlockNumber:    bsc.Commitment.BlockNumber,
				ValidatorSetId: bsc.Commitment.ValidatorSetID,
			}

			gSignatures := make([]ibcgptypes.Signature, len(bsc.Signatures))
			for i, v := range bsc.Signatures {
				gSignatures[i] = ibcgptypes.Signature(v)
			}

			gsc := ibcgptypes.SignedCommitment{
				Commitment: gCommitment,
				Signatures: gSignatures,
			}
			// convert mmrleaf
			var gMMRLeaves []ibcgptypes.MMRLeaf
			t.Logf("gMMRLeaves: %+v", gMMRLeaves)
			leafNum := len(mmrBatchProof.Leaves)
			for i := 0; i < leafNum; i++ {
				leaf := mmrBatchProof.Leaves[i]
				parentNumAndHash := ibcgptypes.ParentNumberAndHash{
					ParentNumber: uint32(leaf.ParentNumberAndHash.ParentNumber),
					ParentHash:   []byte(leaf.ParentNumberAndHash.Hash[:]),
				}
				nextAuthoritySet := ibcgptypes.BeefyAuthoritySet{
					Id:   uint64(leaf.BeefyNextAuthoritySet.ID),
					Len:  uint32(leaf.BeefyNextAuthoritySet.Len),
					Root: []byte(leaf.BeefyNextAuthoritySet.Root[:]),
				}
				parachainHeads := []byte(leaf.ParachainHeads[:])
				gLeaf := ibcgptypes.MMRLeaf{
					Version:               uint32(leaf.Version),
					ParentNumberAndHash:   parentNumAndHash,
					BeefyNextAuthoritySet: nextAuthoritySet,
					ParachainHeads:        parachainHeads,
				}
				t.Logf("gLeaf: %+v", gLeaf)
				gMMRLeaves = append(gMMRLeaves, gLeaf)
			}
			t.Logf("gMMRLeaves: %+v", gMMRLeaves)

			// convert mmr batch proof
			gLeafIndexes := make([]uint64, len(mmrBatchProof.Proof.LeafIndex))
			for i, v := range mmrBatchProof.Proof.LeafIndex {
				gLeafIndexes[i] = uint64(v)
			}

			gProofItems := [][]byte{}
			t.Logf("gProofItems: %+v", gProofItems)
			itemNum := len(mmrBatchProof.Proof.Items)
			for i := 0; i < itemNum; i++ {
				item := mmrBatchProof.Proof.Items[i][:]
				t.Logf("item: %+v", item)
				gProofItems = append(gProofItems, item)
				// gProofItems[i] = item
				t.Logf("gProofItems: %+v", gProofItems)
			}
			t.Logf("gProofItems: %+v", gProofItems)

			gBatchProof := ibcgptypes.MMRBatchProof{
				LeafIndexes: gLeafIndexes,
				LeafCount:   uint64(mmrBatchProof.Proof.LeafCount),
				Items:       gProofItems,
			}

			gMmrLevavesAndProof := ibcgptypes.MMRLeavesAndBatchProof{
				Leaves:        gMMRLeaves,
				MmrBatchProof: gBatchProof,
			}

			// build gBeefyMMR
			gBeefyMMR := ibcgptypes.BeefyMMR{
				SignedCommitment:       gsc,
				SignatureProofs:        authorityProof,
				MmrLeavesAndBatchProof: gMmrLevavesAndProof,
				MmrSize:                mmrSize,
			}

			// step3, build header proof
			// build solochain header map
			solochainHeaderMap, err := beefy.BuildSolochainHeaderMap(localSolochainEndpoint, mmrBatchProof.Proof.LeafIndex)
			require.NoError(t, err)
			t.Logf("solochainHeaderMap: %+v", solochainHeaderMap)

			// verify solochain and proof
			t.Log("\n------------------ VerifySolochainHeader ----------------------------")
			err = beefy.VerifySolochainHeader(mmrBatchProof.Leaves, solochainHeaderMap)
			require.NoError(t, err)
			t.Log("\n------------------ VerifySolochainHeader end ------------------------\n")

			// convert beefy solochain header to pb solochain header
			headerMap := make(map[uint32]ibcgptypes.SolochainHeader)
			for num, header := range solochainHeaderMap {
				headerMap[num] = ibcgptypes.SolochainHeader{
					BlockHeader: header.BlockHeader,
					Timestamp:   ibcgptypes.StateProof(header.Timestamp),
				}
			}

			gSolochainHeaderMap := ibcgptypes.SolochainHeaderMap{
				SolochainHeaderMap: headerMap,
			}

			t.Logf("gSolochainHeaderMap: %+v", gSolochainHeaderMap)

			header_solochainMap := ibcgptypes.Header_SolochainHeaderMap{
				SolochainHeaderMap: &gSolochainHeaderMap,
			}

			// build grandpa pb header
			gheader := ibcgptypes.Header{
				BeefyMmr: gBeefyMMR,
				Message:  &header_solochainMap,
			}
			t.Logf("gpheader: %+v", gheader)

			// mock gp header marshal and unmarshal
			marshalGHeader, err := gheader.Marshal()
			require.NoError(t, err)
			// t.Logf("marshal gHeader: %+v", marshalGHeader)

			t.Log("\n\n------------- mock verify on chain -------------\n")

			// err = gheader.Unmarshal(marshalGHeader)
			var unmarshalGPHeader ibcgptypes.Header
			err = unmarshalGPHeader.Unmarshal(marshalGHeader)
			require.NoError(t, err)
			t.Logf("unmarshal gHeader: %+v", unmarshalGPHeader)

			// step1:verify signature
			// unmarshalBeefyMmr := unmarshalGPHeader.BeefyMmr
			unmarshalBeefyMmr := gheader.BeefyMmr
			t.Logf("gBeefyMMR: %+v", gBeefyMMR)
			t.Logf("unmarshal BeefyMmr: %+v", unmarshalBeefyMmr)

			unmarshalgsc := unmarshalBeefyMmr.SignedCommitment

			rebuildPalyloads := make([]gsrpctypes.PayloadItem, len(unmarshalgsc.Commitment.Payloads))
			// // step1:  verify signature
			for i, v := range unmarshalgsc.Commitment.Payloads {
				rebuildPalyloads[i] = gsrpctypes.PayloadItem{
					ID:   beefy.Bytes2(v.Id),
					Data: v.Data,
				}
			}
			t.Logf("bsc.Commitment.Payload: %+v", bsc.Commitment.Payload)
			t.Logf("gPalyloads: %+v", gPalyloads)
			t.Logf("unmarshal payloads: %+v", unmarshalgsc.Commitment.Payloads)
			t.Logf("rebuildPalyloads: %+v", rebuildPalyloads)

			// convert signature
			rebuildSignatures := make([]beefy.Signature, len(unmarshalgsc.Signatures))
			for i, v := range unmarshalgsc.Signatures {
				rebuildSignatures[i] = beefy.Signature{
					Index:     v.Index,
					Signature: v.Signature,
				}
			}
			t.Logf("bsc.Signatures: %+v", bsc.Signatures)
			t.Logf("gsc.Signatures: %+v", gsc.Signatures)
			t.Logf("unmarshal Signatures: %+v", unmarshalgsc.Signatures)
			t.Logf("rebuildSignatures: %+v", rebuildSignatures)

			// build beefy SignedCommitment
			rebuildBSC := beefy.SignedCommitment{
				Commitment: gsrpctypes.Commitment{
					Payload:        rebuildPalyloads,
					BlockNumber:    unmarshalgsc.Commitment.BlockNumber,
					ValidatorSetID: unmarshalgsc.Commitment.ValidatorSetId,
				},
				Signatures: rebuildSignatures,
			}
			t.Logf("bsc: %+v", bsc)
			t.Logf("gsc: %+v", gsc)
			t.Logf("unmarshalgsc: %+v", unmarshalgsc)
			t.Logf("rebuildBSC: %+v", rebuildBSC)

			t.Log("\n------------------ VerifySignatures --------------------------")
			err = clientState.VerifySignatures(rebuildBSC, unmarshalBeefyMmr.SignatureProofs)
			require.NoError(t, err)
			t.Log("\n------------------ VerifySignatures end ----------------------\n")

			// step2, verify mmr
			// convert mmrleaf
			unmarshalLeaves := unmarshalBeefyMmr.MmrLeavesAndBatchProof.Leaves
			rebuildMMRLeaves := make([]gsrpctypes.MMRLeaf, len(unmarshalLeaves))
			for i, v := range unmarshalLeaves {
				rebuildMMRLeaves[i] = gsrpctypes.MMRLeaf{
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
			t.Logf("mmrBatchProof.Leaves: %+v", mmrBatchProof.Leaves)
			t.Logf("gMMRLeaves: %+v", gMMRLeaves)
			t.Logf("unmarshal mmr Leaves: %+v", unmarshalLeaves)
			t.Logf("rebuildMMRLeaves: %+v", rebuildMMRLeaves)

			// convert mmr batch proof
			unmarshaoLeafIndexes := unmarshalBeefyMmr.MmrLeavesAndBatchProof.MmrBatchProof.LeafIndexes
			rebuildLeafIndexes := make([]gsrpctypes.U64, len(unmarshaoLeafIndexes))
			for i, v := range unmarshaoLeafIndexes {
				rebuildLeafIndexes[i] = gsrpctypes.NewU64(v)
			}
			t.Logf("mmrBatchProof.Proof.LeafIndex: %+v", mmrBatchProof.Proof.LeafIndex)
			t.Logf("gLeafIndexes: %+v", gLeafIndexes)
			t.Logf("unmarshal leafIndexes: %+v", unmarshaoLeafIndexes)
			t.Logf("rebuildLeafIndexes: %+v", rebuildLeafIndexes)

			unmarshalItems := unmarshalBeefyMmr.MmrLeavesAndBatchProof.MmrBatchProof.Items
			rebuildItems := make([]gsrpctypes.H256, len(unmarshalItems))
			for i, v := range unmarshalItems {
				rebuildItems[i] = gsrpctypes.NewH256(v)
			}

			t.Logf("mmrBatchProof.Proof.Items: %+v", mmrBatchProof.Proof.Items)
			t.Logf("gItems: %+v", gProofItems)
			t.Logf("unmarshal proof items: %+v", unmarshalItems)
			t.Logf("rebuildItems: %+v", rebuildItems)

			rebuildMmrBatchProof := beefy.MMRBatchProof{
				LeafIndex: rebuildLeafIndexes,
				LeafCount: gsrpctypes.NewU64(unmarshalBeefyMmr.MmrLeavesAndBatchProof.MmrBatchProof.LeafCount),
				Items:     rebuildItems,
			}

			t.Logf("mmrBatchProof: %+v", mmrBatchProof.Proof)
			t.Logf("gBatchProof: %+v", gBatchProof)
			t.Logf("unmarshal BatchProof: %+v", unmarshalBeefyMmr.MmrLeavesAndBatchProof.MmrBatchProof)
			t.Logf("rebuildMmrBatchProof: %+v", rebuildMmrBatchProof)

			// check mmr height
			require.Less(t, clientState.LatestBeefyHeight, unmarshalBeefyMmr.SignedCommitment.Commitment.BlockNumber)
			t.Log("\n------------------ VerifyMMRBatchProof --------------------------")
			result, err = beefy.VerifyMMRBatchProof(rebuildPalyloads,
				unmarshalBeefyMmr.MmrSize, rebuildMMRLeaves, rebuildMmrBatchProof)
			require.NoError(t, err)
			require.True(t, result)
			t.Log("\n------------------ VerifyMMRBatchProof end ----------------------\n")

			// step3, verify header
			// convert pb solochain header to beefy solochain header
			rebuildSolochainHeaderMap := make(map[uint32]beefy.SolochainHeader)
			unmarshalSolochainHeaderMap := unmarshalGPHeader.GetSolochainHeaderMap()
			for num, header := range unmarshalSolochainHeaderMap.SolochainHeaderMap {
				rebuildSolochainHeaderMap[num] = beefy.SolochainHeader{
					BlockHeader: header.BlockHeader,
					Timestamp:   beefy.StateProof(header.Timestamp),
				}
			}
			t.Logf("solochainHeaderMap: %+v", solochainHeaderMap)
			t.Logf("gSolochainHeaderMap: %+v", gSolochainHeaderMap)
			t.Logf("unmarshal solochainHeaderMap: %+v", *unmarshalSolochainHeaderMap)
			t.Logf("rebuildSolochainHeaderMap: %+v", rebuildSolochainHeaderMap)
			t.Log("\n------------------ VerifySolochainHeader --------------------------")
			err = beefy.VerifySolochainHeader(rebuildMMRLeaves, rebuildSolochainHeaderMap)
			require.NoError(t, err)
			t.Log("\n------------------ VerifySolochainHeader end ----------------------\n")

			// step4, update client state
			clientState.LatestBeefyHeight = latestSignedCommitmentBlockNumber
			clientState.LatestHeight = latestSignedCommitmentBlockNumber
			clientState.MmrRootHash = unmarshalgsc.Commitment.Payloads[0].Data
			// find latest next authority set from mmrleaves
			var latestNextAuthoritySet *ibcgptypes.BeefyAuthoritySet
			var latestAuthoritySetId uint64
			for _, leaf := range unmarshalLeaves {
				if latestAuthoritySetId < leaf.BeefyNextAuthoritySet.Id {
					latestAuthoritySetId = leaf.BeefyNextAuthoritySet.Id
					latestNextAuthoritySet = &leaf.BeefyNextAuthoritySet
				}

			}
			t.Logf("current clientState.AuthoritySet.Id: %+v", clientState.AuthoritySet.Id)
			t.Logf("latestAuthoritySetId: %+v", latestAuthoritySetId)
			t.Logf("latestNextAuthoritySet: %+v", *latestNextAuthoritySet)
			//  update client state authority set
			if clientState.AuthoritySet.Id < latestAuthoritySetId {
				clientState.AuthoritySet = *latestNextAuthoritySet
				t.Logf("update clientState.AuthoritySet : %+v", clientState.AuthoritySet)
			}
			// step5,update consensue state
			for num, header := range unmarshalSolochainHeaderMap.SolochainHeaderMap {
				var decodeHeader gsrpctypes.Header
				err = gsrpccodec.Decode(header.BlockHeader, &decodeHeader)
				require.NoError(t, err)
				var timestamp gsrpctypes.U64
				err = gsrpccodec.Decode(header.Timestamp.Value, &timestamp)
				require.NoError(t, err)

				consensusState := ibcgptypes.ConsensusState{
					Timestamp: time.UnixMilli(int64(timestamp)),
					Root:      decodeHeader.StateRoot[:],
				}
				consensusStateKVStore[num] = consensusState
			}
			t.Logf("latest consensusStateKVStore: %+v", consensusStateKVStore)

			// step6, mock to build and verify state proof
			for num, consnesue := range consensusStateKVStore {
				targetBlockHash, err := localSolochainEndpoint.RPC.Chain.GetBlockHash(uint64(num))
				require.NoError(t, err)
				timestampProof, err := beefy.GetTimestampProof(localSolochainEndpoint, targetBlockHash)
				require.NoError(t, err)

				proofs := make([][]byte, len(timestampProof.Proof))
				for i, proof := range timestampProof.Proof {
					proofs[i] = proof
				}
				t.Logf("timestampProof proofs: %+v", proofs)
				paraTimestampStoragekey := beefy.CreateStorageKeyPrefix("Timestamp", "Now")
				t.Logf("paraTimestampStoragekey: %#x", paraTimestampStoragekey)
				timestampValue := consnesue.Timestamp.UnixMilli()
				t.Logf("timestampValue: %d", timestampValue)
				encodedTimeStampValue, err := gsrpccodec.Encode(timestampValue)
				require.NoError(t, err)
				t.Logf("encodedTimeStampValue: %+v", encodedTimeStampValue)
				err = beefy.VerifyStateProof(proofs, consnesue.Root, paraTimestampStoragekey, encodedTimeStampValue)
				require.NoError(t, err)
				t.Log("beefy.VerifyStateProof(proof,root,key,value) result: True")
			}
			received++

			if received >= 5 {
				return
			}
		case <-timeout:
			t.Logf("timeout reached without getting 2 notifications from subscription")
			return
		}
	}
}

func TestParachainLocalNet(t *testing.T) {
	// t.Skip("todo: get and build data from substrate chain and invoke the grandpa lc to verify data")
	localRelayEndpoint, err := gsrpc.NewSubstrateAPI(beefy.LOCAL_RELAY_ENDPPOIT)
	if err != nil {
		t.Logf("Connecting err: %v", err)
	}
	ch := make(chan interface{})
	sub, err := localRelayEndpoint.Client.Subscribe(
		context.Background(),
		"beefy",
		"subscribeJustifications",
		"unsubscribeJustifications",
		"justifications",
		ch)
	require.NoError(t, err)
	t.Logf("subscribed to %s\n", beefy.LOCAL_RELAY_ENDPPOIT)
	defer sub.Unsubscribe()

	localParachainEndpoint, err := gsrpc.NewSubstrateAPI(beefy.LOCAL_PARACHAIN_ENDPOINT)
	if err != nil {
		t.Logf("Connecting err: %v", err)
	}

	timeout := time.After(24 * time.Hour)
	received := 0
	// var preBlockNumber uint32
	// var preBloackHash gsrpctypes.Hash
	var clientState *ibcgptypes.ClientState
	consensusStateKVStore := make(map[uint32]ibcgptypes.ConsensusState)
	for {
		select {
		case msg := <-ch:
			t.Logf("encoded msg: %s", msg)

			s := &beefy.VersionedFinalityProof{}
			err := gsrpccodec.DecodeFromHex(msg.(string), s)
			if err != nil {
				panic(err)
			}

			t.Logf("decoded msg: %+v\n", s)
			t.Logf("decoded msg: %#v\n", s)
			latestSignedCommitmentBlockNumber := s.SignedCommitment.Commitment.BlockNumber
			// t.Logf("blockNumber: %d\n", latestBlockNumber)
			latestSignedCommitmentBlockHash, err := localRelayEndpoint.RPC.Chain.GetBlockHash(uint64(latestSignedCommitmentBlockNumber))
			require.NoError(t, err)
			t.Logf("latestSignedCommitmentBlockNumber: %d latestSignedCommitmentBlockHash: %#x",
				latestSignedCommitmentBlockNumber, latestSignedCommitmentBlockHash)

			// build and init grandpa lc client state
			if clientState == nil {
				//  init client state
				t.Log("clientState == nil, need to init !")
				authoritySet, err := beefy.GetBeefyAuthoritySet(latestSignedCommitmentBlockHash, localRelayEndpoint, "BeefyAuthorities")
				require.NoError(t, err)
				t.Logf("current authority set: %+v", authoritySet)
				nextAuthoritySet, err := beefy.GetBeefyAuthoritySet(latestSignedCommitmentBlockHash, localRelayEndpoint, "BeefyNextAuthorities")
				require.NoError(t, err)
				t.Logf("next authority set: %+v", nextAuthoritySet)
				// get the parachain latest height
				// the height must be inclueded into relay chain, if not exist ,the height is zero
				fromBlockNumber := latestSignedCommitmentBlockNumber - 7 // for test
				t.Logf("fromBlockNumber: %d toBlockNumber: %d", fromBlockNumber, latestSignedCommitmentBlockNumber)

				fromBlockHash, err := localRelayEndpoint.RPC.Chain.GetBlockHash(uint64(fromBlockNumber))
				require.NoError(t, err)
				t.Logf("fromBlockHash: %#x ", fromBlockHash)
				t.Logf("toBlockHash: %#x", latestSignedCommitmentBlockHash)
				changeSets, err := beefy.QueryParachainStorage(localRelayEndpoint, beefy.LOCAL_PARACHAIN_ID, fromBlockHash, latestSignedCommitmentBlockHash)
				require.NoError(t, err)
				t.Logf("changeSet len: %d", len(changeSets))
				var latestHeight uint32
				if len(changeSets) == 0 {
					latestHeight = 0
				} else {
					var includedParachainHeights []int
					for _, changeSet := range changeSets {
						for _, change := range changeSet.Changes {
							t.Logf("change.StorageKey: %#x", change.StorageKey)
							t.Log("change.HasStorageData: ", change.HasStorageData)
							t.Logf("change.HasStorageData: %#x", change.StorageData)
							if change.HasStorageData {
								var parachainheader gsrpctypes.Header
								// first decode byte array
								var bz []byte
								err = gsrpccodec.Decode(change.StorageData, &bz)
								require.NoError(t, err)
								// second decode header
								err = gsrpccodec.Decode(bz, &parachainheader)
								require.NoError(t, err)
								includedParachainHeights = append(includedParachainHeights, int(parachainheader.Number))
							}
						}

					}
					t.Logf("raw includedParachainHeights: %+v", includedParachainHeights)
					// sort heights and find latest height
					sort.Sort(sort.Reverse(sort.IntSlice(includedParachainHeights)))
					t.Logf("sort.Reverse: %+v", includedParachainHeights)
					latestHeight = uint32(includedParachainHeights[0])
					t.Logf("latestHeight: %d", latestHeight)
				}

				clientState = &ibcgptypes.ClientState{
					ChainId:              beefy.LOCAL_PARACHAIN_ID,
					ChainType:            beefy.CHAINTYPE_PARACHAIN,
					BeefyActivationBlock: beefy.BEEFY_ACTIVATION_BLOCK,
					LatestBeefyHeight:    latestSignedCommitmentBlockNumber,
					MmrRootHash:          s.SignedCommitment.Commitment.Payload[0].Data,
					LatestHeight:         latestHeight,
					FrozenHeight:         0,
					AuthoritySet: ibcgptypes.BeefyAuthoritySet{
						Id:   uint64(authoritySet.ID),
						Len:  uint32(authoritySet.Len),
						Root: authoritySet.Root[:],
					},
					NextAuthoritySet: ibcgptypes.BeefyAuthoritySet{
						Id:   uint64(nextAuthoritySet.ID),
						Len:  uint32(nextAuthoritySet.Len),
						Root: nextAuthoritySet.Root[:],
					},
				}
				t.Logf("init client state: %+v", &clientState)
				t.Logf("init client state: %+v", *clientState)
				// test pb marshal and unmarshal
				marshalCS, err := clientState.Marshal()
				require.NoError(t, err)
				t.Logf("marshal client state: %+v", marshalCS)
				// unmarshal
				// err = clientState.Unmarshal(marshalCS)
				var unmarshalCS ibcgptypes.ClientState
				err = unmarshalCS.Unmarshal(marshalCS)
				require.NoError(t, err)
				t.Logf("unmarshal client state: %+v", unmarshalCS)
				continue
			}

			// step1,build authourities proof for current beefy signatures
			authorities, err := beefy.GetBeefyAuthorities(latestSignedCommitmentBlockHash, localRelayEndpoint, "Authorities")
			require.NoError(t, err)
			bsc := beefy.ConvertCommitment(s.SignedCommitment)
			var authorityIdxes []uint64
			for _, v := range bsc.Signatures {
				idx := v.Index
				authorityIdxes = append(authorityIdxes, uint64(idx))
			}
			_, authorityProof, err := beefy.BuildAuthorityProof(authorities, authorityIdxes)
			require.NoError(t, err)

			// step2,build beefy mmr
			fromBlockNumber := clientState.LatestBeefyHeight + 1

			fromBlockHash, err := localRelayEndpoint.RPC.Chain.GetBlockHash(uint64(fromBlockNumber))
			require.NoError(t, err)

			changeSets, err := beefy.QueryParachainStorage(localRelayEndpoint, beefy.LOCAL_PARACHAIN_ID, fromBlockHash, latestSignedCommitmentBlockHash)
			require.NoError(t, err)

			var targetRelayChainBlockHeights []uint64
			for _, changeSet := range changeSets {
				relayHeader, err := localRelayEndpoint.RPC.Chain.GetHeader(changeSet.Block)
				require.NoError(t, err)

				targetRelayChainBlockHeights = append(targetRelayChainBlockHeights, uint64(relayHeader.Number))

			}

			// build mmr proofs for leaves containing target paraId
			mmrBatchProof, err := beefy.BuildMMRBatchProof(localRelayEndpoint, &latestSignedCommitmentBlockHash, targetRelayChainBlockHeights)
			require.NoError(t, err)

			leafIndex := beefy.ConvertBlockNumberToMmrLeafIndex(uint32(beefy.BEEFY_ACTIVATION_BLOCK), latestSignedCommitmentBlockNumber)
			mmrSize := mmr.LeafIndexToMMRSize(uint64(leafIndex))

			result, err := beefy.VerifyMMRBatchProof(s.SignedCommitment.Commitment.Payload,
				mmrSize, mmrBatchProof.Leaves, mmrBatchProof.Proof)
			require.NoError(t, err)
			require.True(t, result)
			t.Logf("beefy.VerifyMMRBatchProof(s.SignedCommitment.Commitment.Payload, mmrSize,mmrBatchProof.Leaves, mmrBatchProof.Proof) result: %+v", result)

			// convert payloads
			gPalyloads := make([]ibcgptypes.PayloadItem, len(bsc.Commitment.Payload))
			for i, v := range bsc.Commitment.Payload {
				gPalyloads[i] = ibcgptypes.PayloadItem{
					Id:   v.ID[:],
					Data: v.Data,
				}

			}

			t.Logf("bsc.Commitment.Payload: %+v", bsc.Commitment.Payload)
			t.Logf("gpPalyloads: %+v", gPalyloads)

			gCommitment := ibcgptypes.Commitment{
				Payloads:       gPalyloads,
				BlockNumber:    bsc.Commitment.BlockNumber,
				ValidatorSetId: bsc.Commitment.ValidatorSetID,
			}

			gSignatures := make([]ibcgptypes.Signature, len(bsc.Signatures))
			for i, v := range bsc.Signatures {
				gSignatures[i] = ibcgptypes.Signature(v)
			}

			gsc := ibcgptypes.SignedCommitment{
				Commitment: gCommitment,
				Signatures: gSignatures,
			}
			// convert mmrleaf
			var gMMRLeaves []ibcgptypes.MMRLeaf
			t.Logf("gMMRLeaves: %+v", gMMRLeaves)
			leafNum := len(mmrBatchProof.Leaves)
			for i := 0; i < leafNum; i++ {
				leaf := mmrBatchProof.Leaves[i]
				parentNumAndHash := ibcgptypes.ParentNumberAndHash{
					ParentNumber: uint32(leaf.ParentNumberAndHash.ParentNumber),
					ParentHash:   []byte(leaf.ParentNumberAndHash.Hash[:]),
				}
				nextAuthoritySet := ibcgptypes.BeefyAuthoritySet{
					Id:   uint64(leaf.BeefyNextAuthoritySet.ID),
					Len:  uint32(leaf.BeefyNextAuthoritySet.Len),
					Root: []byte(leaf.BeefyNextAuthoritySet.Root[:]),
				}
				parachainHeads := []byte(leaf.ParachainHeads[:])
				gLeaf := ibcgptypes.MMRLeaf{
					Version:               uint32(leaf.Version),
					ParentNumberAndHash:   parentNumAndHash,
					BeefyNextAuthoritySet: nextAuthoritySet,
					ParachainHeads:        parachainHeads,
				}
				t.Logf("gLeaf: %+v", gLeaf)
				gMMRLeaves = append(gMMRLeaves, gLeaf)
			}
			t.Logf("gMMRLeaves: %+v", gMMRLeaves)

			// convert mmr batch proof
			gLeafIndexes := make([]uint64, len(mmrBatchProof.Proof.LeafIndex))
			for i, v := range mmrBatchProof.Proof.LeafIndex {
				gLeafIndexes[i] = uint64(v)
			}

			gProofItems := [][]byte{}
			t.Logf("gProofItems: %+v", gProofItems)
			itemNum := len(mmrBatchProof.Proof.Items)
			for i := 0; i < itemNum; i++ {
				item := mmrBatchProof.Proof.Items[i][:]
				t.Logf("item: %+v", item)
				gProofItems = append(gProofItems, item)
				// gProofItems[i] = item
				t.Logf("gProofItems: %+v", gProofItems)
			}
			t.Logf("gProofItems: %+v", gProofItems)

			gBatchProof := ibcgptypes.MMRBatchProof{
				LeafIndexes: gLeafIndexes,
				LeafCount:   uint64(mmrBatchProof.Proof.LeafCount),
				Items:       gProofItems,
			}

			gMmrLevavesAndProof := ibcgptypes.MMRLeavesAndBatchProof{
				Leaves:        gMMRLeaves,
				MmrBatchProof: gBatchProof,
			}

			// build gBeefyMMR
			gBeefyMMR := ibcgptypes.BeefyMMR{
				SignedCommitment:       gsc,
				SignatureProofs:        authorityProof,
				MmrLeavesAndBatchProof: gMmrLevavesAndProof,
				MmrSize:                mmrSize,
			}

			// TODO: step3, build header proof
			// build parachain header proof and verify that proof
			parachainHeaderMap, err := beefy.BuildParachainHeaderMap(localRelayEndpoint, mmrBatchProof.Proof.LeafIndex, beefy.LOCAL_PARACHAIN_ID)
			require.NoError(t, err)
			t.Logf("parachainHeaderMap: %+v", parachainHeaderMap)

			t.Log("\n----------- VerifyParachainHeader -----------")
			err = beefy.VerifyParachainHeader(mmrBatchProof.Leaves, parachainHeaderMap)
			require.NoError(t, err)
			t.Log("\n------------------------------------------\n")

			// convert beefy parachain header to pb parachain header
			headerMap := make(map[uint32]ibcgptypes.ParachainHeader)
			for num, header := range parachainHeaderMap {
				headerMap[num] = ibcgptypes.ParachainHeader{
					ParachainId: header.ParaId,
					BlockHeader: header.BlockHeader,
					Proofs:      header.Proof,
					HeaderIndex: header.HeaderIndex,
					HeaderCount: header.HeaderCount,
					Timestamp:   ibcgptypes.StateProof(header.Timestamp),
				}
			}

			gParachainHeaderMap := ibcgptypes.ParachainHeaderMap{
				ParachainHeaderMap: headerMap,
			}

			t.Logf("gParachainHeaderMap: %+v", gParachainHeaderMap)

			header_parachainMap := ibcgptypes.Header_ParachainHeaderMap{
				ParachainHeaderMap: &gParachainHeaderMap,
			}

			// build grandpa pb header
			gheader := ibcgptypes.Header{
				BeefyMmr: gBeefyMMR,
				Message:  &header_parachainMap,
			}
			t.Logf("gpheader: %+v", gheader)

			// mock gp header marshal and unmarshal
			marshalGHeader, err := gheader.Marshal()
			require.NoError(t, err)
			// t.Logf("marshal gHeader: %+v", marshalGHeader)

			t.Log("\n------------- mock verify on chain -------------\n")

			// err = gheader.Unmarshal(marshalGHeader)
			var unmarshalGPHeader ibcgptypes.Header
			err = unmarshalGPHeader.Unmarshal(marshalGHeader)
			require.NoError(t, err)
			t.Logf("unmarshal gHeader: %+v", unmarshalGPHeader)

			// verify signature
			// unmarshalBeefyMmr := unmarshalGPHeader.BeefyMmr
			unmarshalBeefyMmr := gheader.BeefyMmr
			t.Logf("gBeefyMMR: %+v", gBeefyMMR)
			t.Logf("unmarshal BeefyMmr: %+v", unmarshalBeefyMmr)

			unmarshalgsc := unmarshalBeefyMmr.SignedCommitment

			rebuildPalyloads := make([]gsrpctypes.PayloadItem, len(unmarshalgsc.Commitment.Payloads))
			// // step1:  verify signature
			for i, v := range unmarshalgsc.Commitment.Payloads {
				rebuildPalyloads[i] = gsrpctypes.PayloadItem{
					ID:   beefy.Bytes2(v.Id),
					Data: v.Data,
				}
			}
			t.Logf("bsc.Commitment.Payload: %+v", bsc.Commitment.Payload)
			t.Logf("gPalyloads: %+v", gPalyloads)
			t.Logf("unmarshal payloads: %+v", unmarshalgsc.Commitment.Payloads)
			t.Logf("rebuildPalyloads: %+v", rebuildPalyloads)

			// convert signature
			rebuildSignatures := make([]beefy.Signature, len(unmarshalgsc.Signatures))
			for i, v := range unmarshalgsc.Signatures {
				rebuildSignatures[i] = beefy.Signature{
					Index:     v.Index,
					Signature: v.Signature,
				}
			}
			t.Logf("bsc.Signatures: %+v", bsc.Signatures)
			t.Logf("gsc.Signatures: %+v", gsc.Signatures)
			t.Logf("unmarshal Signatures: %+v", unmarshalgsc.Signatures)
			t.Logf("rebuildSignatures: %+v", rebuildSignatures)

			// build beefy SignedCommitment
			rebuildBSC := beefy.SignedCommitment{
				Commitment: gsrpctypes.Commitment{
					Payload:        rebuildPalyloads,
					BlockNumber:    unmarshalgsc.Commitment.BlockNumber,
					ValidatorSetID: unmarshalgsc.Commitment.ValidatorSetId,
				},
				Signatures: rebuildSignatures,
			}
			t.Logf("bsc: %+v", bsc)
			t.Logf("gsc: %+v", gsc)
			t.Logf("unmarshalgsc: %+v", unmarshalgsc)
			t.Logf("rebuildBSC: %+v", rebuildBSC)

			t.Log("----------- verify signature -----------")
			err = clientState.VerifySignatures(rebuildBSC, unmarshalBeefyMmr.SignatureProofs)
			require.NoError(t, err)
			t.Log("----------------------------------------")

			// step2, verify mmr
			// convert mmrleaf
			unmarshalLeaves := unmarshalBeefyMmr.MmrLeavesAndBatchProof.Leaves
			rebuildMMRLeaves := make([]gsrpctypes.MMRLeaf, len(unmarshalLeaves))
			for i, v := range unmarshalLeaves {
				rebuildMMRLeaves[i] = gsrpctypes.MMRLeaf{
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
			t.Logf("mmrBatchProof.Leaves: %+v", mmrBatchProof.Leaves)
			t.Logf("gMMRLeaves: %+v", gMMRLeaves)
			t.Logf("unmarshal mmr Leaves: %+v", unmarshalLeaves)
			t.Logf("rebuildMMRLeaves: %+v", rebuildMMRLeaves)

			// convert mmr batch proof
			unmarshalLeafIndexes := unmarshalBeefyMmr.MmrLeavesAndBatchProof.MmrBatchProof.LeafIndexes
			rebuildLeafIndexes := make([]gsrpctypes.U64, len(unmarshalLeafIndexes))
			for i, v := range unmarshalLeafIndexes {
				rebuildLeafIndexes[i] = gsrpctypes.NewU64(v)
			}
			t.Logf("mmrBatchProof.Proof.LeafIndex: %+v", mmrBatchProof.Proof.LeafIndex)
			t.Logf("gLeafIndexes: %+v", gLeafIndexes)
			t.Logf("unmarshal leafIndexes: %+v", unmarshalLeafIndexes)
			t.Logf("rebuildLeafIndexes: %+v", rebuildLeafIndexes)

			unmarshalItmes := unmarshalBeefyMmr.MmrLeavesAndBatchProof.MmrBatchProof.Items
			rebuildItems := make([]gsrpctypes.H256, len(unmarshalItmes))
			for i, v := range unmarshalItmes {
				rebuildItems[i] = gsrpctypes.NewH256(v)
			}

			t.Logf("mmrBatchProof.Proof.Items: %+v", mmrBatchProof.Proof.Items)
			t.Logf("gItems: %+v", gProofItems)
			t.Logf("unmarshal proof items: %+v", unmarshalItmes)
			t.Logf("rebuildItems: %+v", rebuildItems)

			rebuildMmrBatchProof := beefy.MMRBatchProof{
				LeafIndex: rebuildLeafIndexes,
				LeafCount: gsrpctypes.NewU64(unmarshalBeefyMmr.MmrLeavesAndBatchProof.MmrBatchProof.LeafCount),
				Items:     rebuildItems,
			}

			t.Logf("mmrBatchProof: %+v", mmrBatchProof.Proof)
			t.Logf("gBatchProof: %+v", gBatchProof)
			t.Logf("unmarshal BatchProof: %+v", unmarshalBeefyMmr.MmrLeavesAndBatchProof.MmrBatchProof)
			t.Logf("rebuildMmrBatchProof: %+v", rebuildMmrBatchProof)

			// check mmr height
			require.Less(t, clientState.LatestBeefyHeight, unmarshalBeefyMmr.SignedCommitment.Commitment.BlockNumber)

			t.Log("\n----------- verify mmr proof -----------")

			result, err = beefy.VerifyMMRBatchProof(rebuildPalyloads,
				unmarshalBeefyMmr.MmrSize, rebuildMMRLeaves, rebuildMmrBatchProof)
			require.NoError(t, err)
			require.True(t, result)
			t.Log("\n----------------------------------------\n")

			// step3, verify header
			// convert pb parachain header to beefy parachain header
			rebuildParachainHeaderMap := make(map[uint32]beefy.ParachainHeader)
			unmarshalParachainHeaderMap := unmarshalGPHeader.GetParachainHeaderMap()
			for num, header := range unmarshalParachainHeaderMap.ParachainHeaderMap {
				rebuildParachainHeaderMap[num] = beefy.ParachainHeader{
					ParaId:      header.ParachainId,
					BlockHeader: header.BlockHeader,
					Proof:       header.Proofs,
					HeaderIndex: header.HeaderIndex,
					HeaderCount: header.HeaderCount,
					Timestamp:   beefy.StateProof(header.Timestamp),
				}
			}
			t.Logf("parachainHeaderMap: %+v", parachainHeaderMap)
			t.Logf("gParachainHeaderMap: %+v", gParachainHeaderMap)
			t.Logf("unmarshal parachainHeaderMap: %+v", *unmarshalParachainHeaderMap)
			t.Logf("rebuildSolochainHeaderMap: %+v", rebuildParachainHeaderMap)
			t.Log("\n----------- VerifyParachainHeader -----------")
			err = beefy.VerifyParachainHeader(rebuildMMRLeaves, rebuildParachainHeaderMap)
			require.NoError(t, err)
			t.Log("\n--------------------------------------------\n")

			// step4, update client state
			clientState.LatestBeefyHeight = latestSignedCommitmentBlockNumber
			clientState.LatestHeight = latestSignedCommitmentBlockNumber
			clientState.MmrRootHash = unmarshalgsc.Commitment.Payloads[0].Data
			// find latest next authority set from mmrleaves
			var latestNextAuthoritySet *ibcgptypes.BeefyAuthoritySet
			var latestAuthoritySetId uint64
			for _, leaf := range unmarshalLeaves {
				if latestAuthoritySetId < leaf.BeefyNextAuthoritySet.Id {
					latestAuthoritySetId = leaf.BeefyNextAuthoritySet.Id
					latestNextAuthoritySet = &leaf.BeefyNextAuthoritySet
				}

			}
			t.Logf("current clientState.AuthoritySet.Id: %+v", clientState.AuthoritySet.Id)
			t.Logf("latestAuthoritySetId: %+v", latestAuthoritySetId)
			t.Logf("latestNextAuthoritySet: %+v", *latestNextAuthoritySet)
			//  update client state authority set
			if clientState.AuthoritySet.Id < latestAuthoritySetId {
				clientState.AuthoritySet = *latestNextAuthoritySet
				t.Logf("update clientState.AuthoritySet : %+v", clientState.AuthoritySet)
			}
			// step5,update consensue state
			for _, header := range unmarshalParachainHeaderMap.ParachainHeaderMap {
				var decodeHeader gsrpctypes.Header
				err = gsrpccodec.Decode(header.BlockHeader, &decodeHeader)
				require.NoError(t, err)
				var timestamp gsrpctypes.U64
				err = gsrpccodec.Decode(header.Timestamp.Value, &timestamp)
				require.NoError(t, err)

				consensusState := ibcgptypes.ConsensusState{
					Timestamp: time.UnixMilli(int64(timestamp)),
					Root:      decodeHeader.StateRoot[:],
				}
				// note: the block number must be parachain header blocknumber,not relaychain header block number
				consensusStateKVStore[uint32(decodeHeader.Number)] = consensusState
			}
			t.Logf("latest consensusStateKVStore: %+v", consensusStateKVStore)

			// step6, mock to build and verify state proof
			for num, consnesue := range consensusStateKVStore {
				// Note: get data from parachain
				targetBlockHash, err := localParachainEndpoint.RPC.Chain.GetBlockHash(uint64(num))
				require.NoError(t, err)
				timestampProof, err := beefy.GetTimestampProof(localParachainEndpoint, targetBlockHash)
				require.NoError(t, err)

				proofs := make([][]byte, len(timestampProof.Proof))
				for i, proof := range timestampProof.Proof {
					proofs[i] = proof
				}
				t.Logf("timestampProof proofs: %+v", proofs)
				paraTimestampStoragekey := beefy.CreateStorageKeyPrefix("Timestamp", "Now")
				t.Logf("paraTimestampStoragekey: %#x", paraTimestampStoragekey)
				timestampValue := consnesue.Timestamp.UnixMilli()
				t.Logf("timestampValue: %d", timestampValue)
				encodedTimeStampValue, err := gsrpccodec.Encode(timestampValue)
				require.NoError(t, err)
				t.Logf("encodedTimeStampValue: %+v", encodedTimeStampValue)
				t.Log("\n------------- mock verify state proof-------------")

				err = beefy.VerifyStateProof(proofs, consnesue.Root, paraTimestampStoragekey, encodedTimeStampValue)
				require.NoError(t, err)
				t.Log("beefy.VerifyStateProof(proof,root,key,value) result: True")
				t.Log("\n--------------------------------------------------")

			}

			received++

			if received >= 2 {
				return
			}
		case <-timeout:
			t.Logf("timeout reached without getting 2 notifications from subscription")
			return
		}
	}
}

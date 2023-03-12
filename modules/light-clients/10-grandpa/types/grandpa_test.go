package types_test

import (
	"context"
	"sort"
	"testing"
	"time"

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
		ChainId:              "sub-0",
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
					ChainId:              "solosub-0",
					ChainType:            beefy.CHAINTYPE_SOLOCHAIN,
					ParachainId:          0,
					BeefyActivationBlock: beefy.BEEFY_ACTIVATION_BLOCK,
					LatestBeefyHeight:    latestSignedCommitmentBlockNumber,
					MmrRootHash:          s.SignedCommitment.Commitment.Payload[0].Data,
					LatestChainHeight:    latestSignedCommitmentBlockNumber,
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
			t.Logf("bsc: %+v", bsc)
			var authorityIdxes []uint64
			for _, v := range bsc.Signatures {
				idx := v.Index
				authorityIdxes = append(authorityIdxes, uint64(idx))
			}
			_, authorityProof, err := beefy.BuildAuthorityProof(authorities, authorityIdxes)
			require.NoError(t, err)

			// step2,build beefy mmr
			targetHeights := []uint32{uint32(latestSignedCommitmentBlockNumber - 1)}
			// build mmr proofs for leaves containing target paraId
			// mmrBatchProof, err := beefy.BuildMMRBatchProof(localSolochainEndpoint, &latestSignedCommitmentBlockHash, targetHeights)
			mmrBatchProof, err := beefy.BuildMMRProofs(localSolochainEndpoint, targetHeights,
				gsrpctypes.NewOptionU32(gsrpctypes.U32(latestSignedCommitmentBlockNumber)),
				gsrpctypes.NewOptionHashEmpty())
			require.NoError(t, err)
			pbBeefyMMR := ibcgptypes.ToPBBeefyMMR(bsc, mmrBatchProof, authorityProof)
			t.Logf("pbBeefyMMR: %+v", pbBeefyMMR)

			// step3, build header proof
			// build solochain header map
			solochainHeaderMap, err := beefy.BuildSolochainHeaderMap(localSolochainEndpoint, mmrBatchProof.Proof.LeafIndexes)
			require.NoError(t, err)
			t.Logf("solochainHeaderMap: %+v", solochainHeaderMap)

			pbHeader_solochainMap := ibcgptypes.ToPBSolochainHeaderMap(solochainHeaderMap)
			// build grandpa pb header
			pbHeader := ibcgptypes.Header{
				BeefyMmr: pbBeefyMMR,
				Message:  &pbHeader_solochainMap,
			}
			t.Logf("pbHeader: %+v", pbHeader)

			// mock gp header marshal and unmarshal
			marshalPBHeader, err := pbHeader.Marshal()
			require.NoError(t, err)
			t.Logf("marshalPBHeader: %+v", marshalPBHeader)

			t.Log("\n\n------------- mock verify on chain -------------")
			// err = gheader.Unmarshal(marshalGHeader)
			var unmarshalPBHeader ibcgptypes.Header
			err = unmarshalPBHeader.Unmarshal(marshalPBHeader)
			require.NoError(t, err)
			t.Logf("unmarshal gHeader: %+v", unmarshalPBHeader)

			// step1:verify signature
			// unmarshalBeefyMmr := unmarshalGPHeader.BeefyMmr
			unmarshalBeefyMmr := pbHeader.BeefyMmr
			// t.Logf("pbBeefyMMR: %+v", pbBeefyMMR)
			t.Logf("unmarshal BeefyMmr: %+v", unmarshalBeefyMmr)

			unmarshalPBSC := unmarshalBeefyMmr.SignedCommitment
			rebuildBSC := ibcgptypes.ToBeefySC(unmarshalPBSC)
			t.Logf("rebuildBSC: %+v", rebuildBSC)

			t.Log("\n------------------ VerifySignatures --------------------------")
			err = clientState.VerifySignatures(rebuildBSC, unmarshalBeefyMmr.SignatureProofs)
			require.NoError(t, err)
			t.Log("\n------------------ VerifySignatures end ----------------------\n")

			// step2, verify mmr
			rebuildMMRLeaves := ibcgptypes.ToBeefyMMRLeaves(unmarshalBeefyMmr.MmrLeavesAndBatchProof.Leaves)
			t.Logf("rebuildMMRLeaves: %+v", rebuildMMRLeaves)
			rebuildMMRBatchProof := ibcgptypes.ToMMRBatchProof(unmarshalBeefyMmr.MmrLeavesAndBatchProof)
			t.Logf("Convert2MMRBatchProof: %+v", rebuildMMRBatchProof)
			// check mmr height
			require.Less(t, clientState.LatestBeefyHeight, unmarshalBeefyMmr.SignedCommitment.Commitment.BlockNumber)
			t.Log("\n------------------ VerifyMMRBatchProof --------------------------")
			result, err := beefy.VerifyMMRBatchProof(rebuildBSC.Commitment.Payload,
				unmarshalBeefyMmr.MmrSize, rebuildMMRLeaves, rebuildMMRBatchProof)
			require.NoError(t, err)
			require.True(t, result)
			t.Log("\n------------------ VerifyMMRBatchProof end ----------------------\n")

			// step3, verify header
			// convert pb solochain header to beefy solochain header
			rebuildSolochainHeaderMap := make(map[uint32]beefy.SolochainHeader)
			unmarshalSolochainHeaderMap := unmarshalPBHeader.GetSolochainHeaderMap()
			for num, header := range unmarshalSolochainHeaderMap.SolochainHeaderMap {
				rebuildSolochainHeaderMap[num] = beefy.SolochainHeader{
					BlockHeader: header.BlockHeader,
					Timestamp:   beefy.StateProof(header.Timestamp),
				}
			}
			// t.Logf("rebuildSolochainHeaderMap: %+v", rebuildSolochainHeaderMap)
			t.Logf("unmarshal solochainHeaderMap: %+v", *unmarshalSolochainHeaderMap)

			t.Log("\n------------------ VerifySolochainHeader --------------------------")
			// err = beefy.VerifySolochainHeader(rebuildMMRLeaves, rebuildSolochainHeaderMap)
			// require.NoError(t, err)
			err = clientState.VerifyHeader(unmarshalPBHeader, rebuildMMRLeaves)
			require.NoError(t, err)
			t.Log("\n------------------ VerifySolochainHeader end ----------------------\n")

			// step4, update client state
			// update client height
			clientState.LatestBeefyHeight = latestSignedCommitmentBlockNumber
			clientState.LatestChainHeight = latestSignedCommitmentBlockNumber
			clientState.MmrRootHash = unmarshalPBSC.Commitment.Payloads[0].Data
			// find latest next authority set from mmrleaves
			var latestNextAuthoritySet *ibcgptypes.BeefyAuthoritySet
			var latestAuthoritySetId uint64
			for _, leaf := range rebuildMMRLeaves {
				if latestAuthoritySetId < uint64(leaf.BeefyNextAuthoritySet.ID) {
					latestAuthoritySetId = uint64(leaf.BeefyNextAuthoritySet.ID)
					latestNextAuthoritySet = &ibcgptypes.BeefyAuthoritySet{
						Id:   uint64(leaf.BeefyNextAuthoritySet.ID),
						Len:  uint32(leaf.BeefyNextAuthoritySet.Len),
						Root: leaf.BeefyNextAuthoritySet.Root[:],
					}
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
			var latestHeight uint32
			for _, header := range unmarshalSolochainHeaderMap.SolochainHeaderMap {
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
				consensusStateKVStore[uint32(decodeHeader.Number)] = consensusState

				if latestHeight < uint32(decodeHeader.Number) {
					latestHeight = uint32(decodeHeader.Number)
				}
			}
			t.Logf("latest consensusStateKVStore: %+v", consensusStateKVStore)
			t.Logf("latest height and consensus state: %d,%+v", latestHeight, consensusStateKVStore[latestHeight])

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
				t.Logf("timestampStoragekey: %#x", paraTimestampStoragekey)
				timestampValue := consnesue.Timestamp.UnixMilli()
				t.Logf("timestampValue: %d", timestampValue)
				encodedTimeStampValue, err := gsrpccodec.Encode(timestampValue)
				require.NoError(t, err)
				t.Logf("encodedTimeStampValue: %+v", encodedTimeStampValue)
				t.Log("\n------------------ VerifyStateProof --------------------------")
				err = beefy.VerifyStateProof(proofs, consnesue.Root, paraTimestampStoragekey, encodedTimeStampValue)
				require.NoError(t, err)
				t.Log("beefy.VerifyStateProof(proof,root,key,value) result: True")
				t.Log("\n------------------ VerifyStateProof end ----------------------\n")
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
	t.Logf("subscribed to relaychain %s\n", beefy.LOCAL_RELAY_ENDPPOIT)
	defer sub.Unsubscribe()

	localParachainEndpoint, err := gsrpc.NewSubstrateAPI(beefy.LOCAL_PARACHAIN_ENDPOINT)
	if err != nil {
		t.Logf("Connecting err: %v", err)
	}
	t.Logf("subscribed to parachain %s\n", beefy.LOCAL_RELAY_ENDPPOIT)

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
				var latestChainHeight uint32
				if len(changeSets) == 0 {
					latestChainHeight = 0
				} else {
					var packedParachainHeights []int
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
								packedParachainHeights = append(packedParachainHeights, int(parachainheader.Number))
							}
						}

					}
					t.Logf("raw packedParachainHeights: %+v", packedParachainHeights)
					// sort heights and find latest height
					sort.Sort(sort.Reverse(sort.IntSlice(packedParachainHeights)))
					t.Logf("sort.Reverse: %+v", packedParachainHeights)
					latestChainHeight = uint32(packedParachainHeights[0])
					t.Logf("latestHeight: %d", latestChainHeight)
				}

				clientState = &ibcgptypes.ClientState{
					ChainId:              "parachain-1",
					ChainType:            beefy.CHAINTYPE_PARACHAIN,
					ParachainId:          beefy.LOCAL_PARACHAIN_ID,
					BeefyActivationBlock: beefy.BEEFY_ACTIVATION_BLOCK,
					LatestBeefyHeight:    latestSignedCommitmentBlockNumber,
					MmrRootHash:          s.SignedCommitment.Commitment.Payload[0].Data,
					LatestChainHeight:    latestChainHeight,
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

			var targetRelaychainBlockHeights []uint32
			for _, changeSet := range changeSets {
				relayHeader, err := localRelayEndpoint.RPC.Chain.GetHeader(changeSet.Block)
				require.NoError(t, err)
				targetRelaychainBlockHeights = append(targetRelaychainBlockHeights, uint32(relayHeader.Number))
			}

			// build mmr proofs for leaves containing target paraId
			// mmrBatchProof, err := beefy.BuildMMRBatchProof(localRelayEndpoint, &latestSignedCommitmentBlockHash, targetRelaychainBlockHeights)
			mmrBatchProof, err := beefy.BuildMMRProofs(localRelayEndpoint, targetRelaychainBlockHeights,
				gsrpctypes.NewOptionU32(gsrpctypes.U32(latestSignedCommitmentBlockNumber)),
				gsrpctypes.NewOptionHashEmpty())
			require.NoError(t, err)

			pbBeefyMMR := ibcgptypes.ToPBBeefyMMR(bsc, mmrBatchProof, authorityProof)
			t.Logf("pbBeefyMMR: %+v", pbBeefyMMR)

			// step3, build header proof
			// build parachain header proof and verify that proof
			parachainHeaderMap, err := beefy.BuildParachainHeaderMap(localRelayEndpoint, localParachainEndpoint,
				mmrBatchProof.Proof.LeafIndexes, beefy.LOCAL_PARACHAIN_ID)
			require.NoError(t, err)
			t.Logf("parachainHeaderMap: %+v", parachainHeaderMap)

			// convert beefy parachain header to pb parachain header
			pbHeader_parachainMap := ibcgptypes.ToPBParachainHeaderMap(parachainHeaderMap)
			t.Logf("pbHeader_parachainMap: %+v", pbHeader_parachainMap)

			// build grandpa pb header
			pbHeader := ibcgptypes.Header{
				BeefyMmr: pbBeefyMMR,
				Message:  &pbHeader_parachainMap,
			}
			t.Logf("gpheader: %+v", pbHeader)

			// mock gp header marshal and unmarshal
			marshalPBHeader, err := pbHeader.Marshal()
			require.NoError(t, err)
			// t.Logf("marshal gHeader: %+v", marshalGHeader)

			t.Log("\n------------- mock verify on chain -------------\n")

			// err = gheader.Unmarshal(marshalGHeader)
			var unmarshalPBHeader ibcgptypes.Header
			err = unmarshalPBHeader.Unmarshal(marshalPBHeader)
			require.NoError(t, err)
			t.Logf("unmarshal gHeader: %+v", unmarshalPBHeader)

			// verify signature
			// unmarshalBeefyMmr := unmarshalGPHeader.BeefyMmr
			unmarshalBeefyMmr := pbHeader.BeefyMmr
			// t.Logf("gBeefyMMR: %+v", gBeefyMMR)
			t.Logf("unmarshal BeefyMmr: %+v", unmarshalBeefyMmr)

			unmarshalPBSC := unmarshalBeefyMmr.SignedCommitment

			rebuildBSC := ibcgptypes.ToBeefySC(unmarshalPBSC)
			t.Logf("rebuildBSC: %+v", rebuildBSC)

			t.Log("----------- verify signature -----------")
			err = clientState.VerifySignatures(rebuildBSC, unmarshalBeefyMmr.SignatureProofs)
			require.NoError(t, err)
			t.Log("----------------------------------------")

			// step2, verify mmr
			rebuildMMRLeaves := ibcgptypes.ToBeefyMMRLeaves(unmarshalBeefyMmr.MmrLeavesAndBatchProof.Leaves)
			t.Logf("rebuildMMRLeaves: %+v", rebuildMMRLeaves)
			beefyMmrBatchProof := ibcgptypes.ToMMRBatchProof(unmarshalBeefyMmr.MmrLeavesAndBatchProof)
			t.Logf("Convert2MMRBatchProof: %+v", beefyMmrBatchProof)

			// check mmr height
			require.Less(t, clientState.LatestBeefyHeight, unmarshalBeefyMmr.SignedCommitment.Commitment.BlockNumber)

			t.Log("\n----------- verify mmr proof -----------")
			result, err := beefy.VerifyMMRBatchProof(rebuildBSC.Commitment.Payload,
				unmarshalBeefyMmr.MmrSize, rebuildMMRLeaves, beefyMmrBatchProof)
			require.NoError(t, err)
			require.True(t, result)
			t.Log("\n----------------------------------------\n")

			// step3, verify header
			// convert pb parachain header to beefy parachain header
			rebuildParachainHeaderMap := make(map[uint32]beefy.ParachainHeader)
			unmarshalParachainHeaderMap := unmarshalPBHeader.GetParachainHeaderMap()
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
			// t.Logf("parachainHeaderMap: %+v", parachainHeaderMap)
			// t.Logf("gParachainHeaderMap: %+v", gParachainHeaderMap)
			t.Logf("unmarshal parachainHeaderMap: %+v", *unmarshalParachainHeaderMap)
			// t.Logf("rebuildSolochainHeaderMap: %+v", rebuildParachainHeaderMap)
			t.Log("\n----------- VerifyParachainHeader -----------")
			// err = beefy.VerifyParachainHeader(rebuildMMRLeaves, rebuildParachainHeaderMap)
			// require.NoError(t, err)
			err = clientState.VerifyHeader(unmarshalPBHeader, rebuildMMRLeaves)
			require.NoError(t, err)
			t.Log("\n--------------------------------------------\n")

			// step4, update client state
			clientState.LatestBeefyHeight = latestSignedCommitmentBlockNumber
			clientState.LatestChainHeight = latestSignedCommitmentBlockNumber
			clientState.MmrRootHash = unmarshalPBSC.Commitment.Payloads[0].Data
			// find latest next authority set from mmrleaves
			var latestNextAuthoritySet *ibcgptypes.BeefyAuthoritySet
			var latestAuthoritySetId uint64
			for _, leaf := range rebuildMMRLeaves {
				if latestAuthoritySetId < uint64(leaf.BeefyNextAuthoritySet.ID) {
					latestAuthoritySetId = uint64(leaf.BeefyNextAuthoritySet.ID)
					latestNextAuthoritySet = &ibcgptypes.BeefyAuthoritySet{
						Id:   uint64(leaf.BeefyNextAuthoritySet.ID),
						Len:  uint32(leaf.BeefyNextAuthoritySet.Len),
						Root: leaf.BeefyNextAuthoritySet.Root[:],
					}
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

			var latestHeight uint32
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
				if latestHeight < uint32(decodeHeader.Number) {
					latestHeight = uint32(decodeHeader.Number)
				}
			}
			t.Logf("latest consensusStateKVStore: %+v", consensusStateKVStore)
			t.Logf("latest height and consensus state: %d,%+v", latestHeight, consensusStateKVStore[latestHeight])

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
			if received >= 5 {
				return
			}
		case <-timeout:
			t.Logf("timeout reached without getting 2 notifications from subscription")
			return
		}
	}
}

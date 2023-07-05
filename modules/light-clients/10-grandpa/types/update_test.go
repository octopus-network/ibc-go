package types_test

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/ComposableFi/go-merkle-trees/mmr"
	"github.com/octopus-network/beefy-go/beefy"
	tmtypes "github.com/tendermint/tendermint/types"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	gsrpctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	gsrpccodec "github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v6/modules/core/23-commitment/types"
	types "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	ibcgptypes "github.com/cosmos/ibc-go/v6/modules/light-clients/10-grandpa/types"
	ibctestingmock "github.com/cosmos/ibc-go/v6/testing/mock"
)

func (suite *GrandpaTestSuite) TestCheckHeaderAndUpdateState() {
	var (
		clientState     *types.ClientState
		consensusState  *types.ConsensusState
		consStateHeight clienttypes.Height
		newHeader       *types.Header
		currentTime     time.Time
		bothValSet      *tmtypes.ValidatorSet
		bothSigners     map[string]tmtypes.PrivValidator
	)

	// Setup different validators and signers for testing different types of updates
	altPrivVal := ibctestingmock.NewPV()
	altPubKey, err := altPrivVal.GetPubKey()
	suite.Require().NoError(err)

	revisionHeight := int64(height.RevisionHeight)

	// create modified heights to use for test-cases
	heightPlus1 := clienttypes.NewHeight(height.RevisionNumber, height.RevisionHeight+1)
	heightMinus1 := clienttypes.NewHeight(height.RevisionNumber, height.RevisionHeight-1)
	heightMinus3 := clienttypes.NewHeight(height.RevisionNumber, height.RevisionHeight-3)
	heightPlus5 := clienttypes.NewHeight(height.RevisionNumber, height.RevisionHeight+5)

	altVal := tmtypes.NewValidator(altPubKey, revisionHeight)
	// Create alternative validator set with only altVal, invalid update (too much change in valSet)
	altValSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{altVal})
	altSigners := getAltSigners(altVal, altPrivVal)

	testCases := []struct {
		name      string
		setup     func(*GrandpaTestSuite)
		expFrozen bool
		expPass   bool
	}{
		{
			name: "successful update with next height and same validator set",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.headerTime, suite.valSet, suite.valSet, suite.valSet, suite.signers)
				currentTime = suite.now
			},
			expFrozen: false,
			expPass:   true,
		},
		{
			name: "successful update with future height and different validator set",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus5.RevisionHeight), height, suite.headerTime, bothValSet, bothValSet, suite.valSet, bothSigners)
				currentTime = suite.now
			},
			expFrozen: false,
			expPass:   true,
		},
		{
			name: "successful update with next height and different validator set",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), bothValSet.Hash())
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.headerTime, bothValSet, bothValSet, bothValSet, bothSigners)
				currentTime = suite.now
			},
			expFrozen: false,
			expPass:   true,
		},
		{
			name: "successful update for a previous height",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				consStateHeight = heightMinus3
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightMinus1.RevisionHeight), heightMinus3, suite.headerTime, bothValSet, bothValSet, suite.valSet, bothSigners)
				currentTime = suite.now
			},
			expFrozen: false,
			expPass:   true,
		},
		{
			name: "successful update for a previous revision",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainIDRevision1, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				consStateHeight = heightMinus3
				newHeader = suite.chainA.CreateTMClientHeader(chainIDRevision0, int64(height.RevisionHeight), heightMinus3, suite.headerTime, bothValSet, bothValSet, suite.valSet, bothSigners)
				currentTime = suite.now
			},
			expPass: true,
		},
		{
			name: "successful update with identical header to a previous update",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, heightPlus1, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.headerTime, suite.valSet, suite.valSet, suite.valSet, suite.signers)
				currentTime = suite.now
				ctx := suite.chainA.GetContext().WithBlockTime(currentTime)
				// Store the header's consensus state in client store before UpdateClient call
				suite.chainA.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(ctx, clientID, heightPlus1, newHeader.ConsensusState())
			},
			expFrozen: false,
			expPass:   true,
		},
		{
			name: "misbehaviour detection: header conflicts with existing consensus state",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, heightPlus1, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.headerTime, suite.valSet, suite.valSet, suite.valSet, suite.signers)
				currentTime = suite.now
				ctx := suite.chainA.GetContext().WithBlockTime(currentTime)
				// Change the consensus state of header and store in client store to create a conflict
				conflictConsState := newHeader.ConsensusState()
				conflictConsState.Root = commitmenttypes.NewMerkleRoot([]byte("conflicting apphash"))
				suite.chainA.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(ctx, clientID, heightPlus1, conflictConsState)
			},
			expFrozen: true,
			expPass:   true,
		},
		{
			name: "misbehaviour detection: previous consensus state time is not before header time. time monotonicity violation",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				// create an intermediate consensus state with the same time as the newHeader to create a time violation.
				// header time is after client time
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus5.RevisionHeight), height, suite.headerTime, suite.valSet, suite.valSet, suite.valSet, suite.signers)
				currentTime = suite.now
				prevConsensusState := types.NewConsensusState(suite.headerTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				ctx := suite.chainA.GetContext().WithBlockTime(currentTime)
				suite.chainA.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(ctx, clientID, heightPlus1, prevConsensusState)
				clientStore := suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(ctx, clientID)
				types.SetIterationKey(clientStore, heightPlus1)
			},
			expFrozen: true,
			expPass:   true,
		},
		{
			name: "misbehaviour detection: next consensus state time is not after header time. time monotonicity violation",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				// create the next consensus state with the same time as the intermediate newHeader to create a time violation.
				// header time is after clientTime
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.headerTime, suite.valSet, suite.valSet, suite.valSet, suite.signers)
				currentTime = suite.now
				nextConsensusState := types.NewConsensusState(suite.headerTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				ctx := suite.chainA.GetContext().WithBlockTime(currentTime)
				suite.chainA.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(ctx, clientID, heightPlus5, nextConsensusState)
				clientStore := suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(ctx, clientID)
				types.SetIterationKey(clientStore, heightPlus5)
			},
			expFrozen: true,
			expPass:   true,
		},
		{
			name: "unsuccessful update with incorrect header chain-id",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader("ethermint", int64(heightPlus1.RevisionHeight), height, suite.headerTime, suite.valSet, suite.valSet, suite.valSet, suite.signers)
				currentTime = suite.now
			},
			expFrozen: false,
			expPass:   false,
		},
		{
			name: "unsuccessful update to a future revision",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainIDRevision0, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainIDRevision1, 1, height, suite.headerTime, suite.valSet, suite.valSet, suite.valSet, suite.signers)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "unsuccessful update: header height revision and trusted height revision mismatch",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainIDRevision1, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, clienttypes.NewHeight(1, 1), commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainIDRevision1, 3, height, suite.headerTime, suite.valSet, suite.valSet, suite.valSet, suite.signers)
				currentTime = suite.now
			},
			expFrozen: false,
			expPass:   false,
		},
		{
			name: "unsuccessful update with next height: update header mismatches nextValSetHash",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.headerTime, bothValSet, bothValSet, suite.valSet, bothSigners)
				currentTime = suite.now
			},
			expFrozen: false,
			expPass:   false,
		},
		{
			name: "unsuccessful update with next height: update header mismatches different nextValSetHash",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), bothValSet.Hash())
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.headerTime, suite.valSet, suite.valSet, bothValSet, suite.signers)
				currentTime = suite.now
			},
			expFrozen: false,
			expPass:   false,
		},
		{
			name: "unsuccessful update with future height: too much change in validator set",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus5.RevisionHeight), height, suite.headerTime, altValSet, altValSet, suite.valSet, altSigners)
				currentTime = suite.now
			},
			expFrozen: false,
			expPass:   false,
		},
		{
			name: "unsuccessful updates, passed in incorrect trusted validators for given consensus state",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus5.RevisionHeight), height, suite.headerTime, bothValSet, bothValSet, bothValSet, bothSigners)
				currentTime = suite.now
			},
			expFrozen: false,
			expPass:   false,
		},
		{
			name: "unsuccessful update: trusting period has passed since last client timestamp",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.headerTime, suite.valSet, suite.valSet, suite.valSet, suite.signers)
				// make current time pass trusting period from last timestamp on clientstate
				currentTime = suite.now.Add(trustingPeriod)
			},
			expFrozen: false,
			expPass:   false,
		},
		{
			name: "unsuccessful update: header timestamp is past current timestamp",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.now.Add(time.Minute), suite.valSet, suite.valSet, suite.valSet, suite.signers)
				currentTime = suite.now
			},
			expFrozen: false,
			expPass:   false,
		},
		{
			name: "unsuccessful update: header timestamp is not past last client timestamp",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.clientTime, suite.valSet, suite.valSet, suite.valSet, suite.signers)
				currentTime = suite.now
			},
			expFrozen: false,
			expPass:   false,
		},
		{
			name: "header basic validation failed",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightPlus1.RevisionHeight), height, suite.headerTime, suite.valSet, suite.valSet, suite.valSet, suite.signers)
				// cause new header to fail validatebasic by changing commit height to mismatch header height
				newHeader.SignedHeader.Commit.Height = revisionHeight - 1
				currentTime = suite.now
			},
			expFrozen: false,
			expPass:   false,
		},
		{
			name: "header height < consensus height",
			setup: func(suite *GrandpaTestSuite) {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, clienttypes.NewHeight(height.RevisionNumber, heightPlus5.RevisionHeight), commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), suite.valsHash)
				// Make new header at height less than latest client state
				newHeader = suite.chainA.CreateTMClientHeader(chainID, int64(heightMinus1.RevisionHeight), height, suite.headerTime, suite.valSet, suite.valSet, suite.valSet, suite.signers)
				currentTime = suite.now
			},
			expFrozen: false,
			expPass:   false,
		},
	}

	for i, tc := range testCases {
		tc := tc
		suite.Run(fmt.Sprintf("Case: %s", tc.name), func() {
			suite.SetupTest() // reset metadata writes
			// Create bothValSet with both suite validator and altVal. Would be valid update
			bothValSet, bothSigners = getBothSigners(suite, altVal, altPrivVal)

			consStateHeight = height // must be explicitly changed
			// setup test
			tc.setup(suite)

			// Set current timestamp in context
			ctx := suite.chainA.GetContext().WithBlockTime(currentTime)

			// Set trusted consensus state in client store
			suite.chainA.App.GetIBCKeeper().ClientKeeper.SetClientConsensusState(ctx, clientID, consStateHeight, consensusState)

			height := newHeader.GetHeight()
			expectedConsensus := &types.ConsensusState{
				Timestamp:          newHeader.GetTime(),
				Root:               commitmenttypes.NewMerkleRoot(newHeader.Header.GetAppHash()),
				NextValidatorsHash: newHeader.Header.NextValidatorsHash,
			}

			newClientState, consensusState, err := clientState.CheckHeaderAndUpdateState(
				ctx,
				suite.cdc,
				suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), clientID), // pass in clientID prefixed clientStore
				newHeader,
			)

			if tc.expPass {
				suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)

				suite.Require().Equal(tc.expFrozen, !newClientState.(*types.ClientState).FrozenHeight.IsZero(), "client state status is unexpected after update")

				// further writes only happen if update is not misbehaviour
				if !tc.expFrozen {
					// Determine if clientState should be updated or not
					// TODO: check the entire Height struct once GetLatestHeight returns clienttypes.Height
					if height.GT(clientState.LatestHeight) {
						// Header Height is greater than clientState latest Height, clientState should be updated with header.GetHeight()
						suite.Require().Equal(height, newClientState.GetLatestHeight(), "clientstate height did not update")
					} else {
						// Update will add past consensus state, clientState should not be updated at all
						suite.Require().Equal(clientState.LatestHeight, newClientState.GetLatestHeight(), "client state height updated for past header")
					}

					suite.Require().Equal(expectedConsensus, consensusState, "valid test case %d failed: %s", i, tc.name)
				}
			} else {
				suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
				suite.Require().Nil(newClientState, "invalid test case %d passed: %s", i, tc.name)
				suite.Require().Nil(consensusState, "invalid test case %d passed: %s", i, tc.name)
			}
		})
	}
}

// This test case need to spin up a substrate chain
func (suite *GrandpaTestSuite) TestSubchainLocalNet() {
	suite.Suite.T().Skip("need to setup relaychain or subchain")
	localSubchainEndpoint, err := gsrpc.NewSubstrateAPI(beefy.LOCAL_RELAY_ENDPPOIT)
	if err != nil {
		suite.Suite.T().Logf("Connecting err: %v", err)
	}
	ch := make(chan interface{})
	sub, err := localSubchainEndpoint.Client.Subscribe(
		context.Background(),
		"beefy",
		"subscribeJustifications",
		"unsubscribeJustifications",
		"justifications",
		ch)
	suite.Require().NoError(err)
	suite.Suite.T().Logf("subscribed to %s\n", beefy.LOCAL_RELAY_ENDPPOIT)

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
			suite.Suite.T().Logf("encoded msg: %s", msg)

			s := &beefy.VersionedFinalityProof{}
			err := gsrpccodec.DecodeFromHex(msg.(string), s)
			if err != nil {
				panic(err)
			}

			suite.Suite.T().Logf("decoded msg: %+v\n", s)
			// suite.Suite.T().Logf("decoded msg: %#v\n", s)
			latestSignedCommitmentBlockNumber := s.SignedCommitment.Commitment.BlockNumber
			// suite.Suite.T().Logf("blockNumber: %d\n", latestBlockNumber)
			latestSignedCommitmentBlockHash, err := localSubchainEndpoint.RPC.Chain.GetBlockHash(uint64(latestSignedCommitmentBlockNumber))
			suite.Require().NoError(err)
			suite.Suite.T().Logf("latestSignedCommitmentBlockNumber: %d latestSignedCommitmentBlockHash: %#x",
				latestSignedCommitmentBlockNumber, latestSignedCommitmentBlockHash)

			// build and init grandpa lc client state
			if clientState == nil {
				//  init client state
				suite.Suite.T().Log("clientState == nil, need to init !")
				authoritySet, err := beefy.GetBeefyAuthoritySet(latestSignedCommitmentBlockHash, localSubchainEndpoint, "BeefyAuthorities")
				suite.Require().NoError(err)
				suite.Suite.T().Logf("current authority set: %+v", authoritySet)
				nextAuthoritySet, err := beefy.GetBeefyAuthoritySet(latestSignedCommitmentBlockHash, localSubchainEndpoint, "BeefyNextAuthorities")
				suite.Require().NoError(err)
				suite.Suite.T().Logf("next authority set: %+v", nextAuthoritySet)
				// get the parachain or subchain latest height
				// the height must be inclueded into relay chain, if not exist ,the height is zero
				fromBlockNumber := latestSignedCommitmentBlockNumber - 7 // for test
				suite.Suite.T().Logf("fromBlockNumber: %d toBlockNumber: %d", fromBlockNumber, latestSignedCommitmentBlockNumber)

				fromBlockHash, err := localSubchainEndpoint.RPC.Chain.GetBlockHash(uint64(fromBlockNumber))
				suite.Require().NoError(err)
				suite.Suite.T().Logf("fromBlockHash: %#x ", fromBlockHash)
				suite.Suite.T().Logf("toBlockHash: %#x", latestSignedCommitmentBlockHash)

				clientState = &ibcgptypes.ClientState{
					ChainId:           "sub-0",
					ChainType:         beefy.CHAINTYPE_SUBCHAIN,
					ParachainId:       0,
					LatestBeefyHeight: clienttypes.NewHeight(clienttypes.ParseChainID("sub-0"), uint64(latestSignedCommitmentBlockNumber)),
					LatestMmrRoot:     s.SignedCommitment.Commitment.Payload[0].Data,
					LatestChainHeight: clienttypes.NewHeight(clienttypes.ParseChainID("sub-0"), uint64(latestSignedCommitmentBlockNumber)),
					FrozenHeight:      clienttypes.NewHeight(clienttypes.ParseChainID("sub-0"), 0),
					LatestAuthoritySet: ibcgptypes.BeefyAuthoritySet{
						Id:   uint64(authoritySet.ID),
						Len:  uint32(authoritySet.Len),
						Root: authoritySet.Root[:],
					},
				}
				// suite.Suite.T().Logf("init client state: %+v", &clientState)
				suite.Suite.T().Logf("init client state: %+v", *clientState)
				// test pb marshal and unmarshal
				marshalCS, err := clientState.Marshal()
				suite.Require().NoError(err)
				suite.Suite.T().Logf("marshal client state: %+v", marshalCS)
				// unmarshal
				// err = clientState.Unmarshal(marshalCS)
				var unmarshalCS ibcgptypes.ClientState
				err = unmarshalCS.Unmarshal(marshalCS)
				suite.Require().NoError(err)
				suite.Suite.T().Logf("unmarshal client state: %+v", unmarshalCS)
				continue
			}

			// step1,build authourities proof for current beefy signatures
			authorities, err := beefy.GetBeefyAuthorities(latestSignedCommitmentBlockHash, localSubchainEndpoint, "Authorities")
			suite.Require().NoError(err)
			bsc := beefy.ConvertCommitment(s.SignedCommitment)
			suite.Suite.T().Logf("bsc: %+v", bsc)
			var authorityIdxes []uint64
			for _, v := range bsc.Signatures {
				idx := v.Index
				authorityIdxes = append(authorityIdxes, uint64(idx))
			}
			_, authorityProof, err := beefy.BuildAuthorityProof(authorities, authorityIdxes)
			suite.Require().NoError(err)

			// step2,build beefy mmr
			// targetHeights := []uint32{uint32(latestSignedCommitmentBlockNumber - 1)}
			targetHeights := []uint32{uint32(latestSignedCommitmentBlockNumber)}
			// build mmr proofs for leaves containing target paraId
			// beefyMmrProof, err := beefy.BuildMMRBatchProof(localSolochainEndpoint, &latestSignedCommitmentBlockHash, targetHeights)
			beefyMmrProof, err := beefy.BuildMMRProofs(localSubchainEndpoint, targetHeights,
				gsrpctypes.NewOptionU32(gsrpctypes.U32(latestSignedCommitmentBlockNumber)),
				gsrpctypes.NewOptionHashEmpty())
			suite.Require().NoError(err)
			pbBeefyMMR := ibcgptypes.ToPBBeefyMMR(bsc, beefyMmrProof, authorityProof)
			suite.Suite.T().Logf("pbBeefyMMR: %+v", pbBeefyMMR)

			// step3, build header proof
			// build subchain header
			// mock targetheights
			targetHeights = append(targetHeights, uint32(latestSignedCommitmentBlockNumber-1))
			suite.Suite.T().Logf("targetHeights: %+v", targetHeights)
			subChainHeaderMmrProof, err := beefy.BuildMMRProofs(localSubchainEndpoint, targetHeights,
				gsrpctypes.NewOptionU32(gsrpctypes.U32(latestSignedCommitmentBlockNumber)),
				gsrpctypes.NewOptionHashEmpty())
			suite.Require().NoError(err)
			subchainHeaders, err := beefy.BuildSubchainHeaders(localSubchainEndpoint, subChainHeaderMmrProof.Proof.LeafIndexes, "sub-0")
			suite.Require().NoError(err)
			suite.Suite.T().Logf("subchainHeaders: %+v", subchainHeaders)
			// suite.Suite.T().Logf("subchainHeaders: %#v", subchainHeaders)

			pbHeader_subchainHeaders := ibcgptypes.ToPBSubchainHeaders(subchainHeaders, subChainHeaderMmrProof)
			// build grandpa pb header
			pbHeader := ibcgptypes.Header{
				BeefyMmr: &pbBeefyMMR,
				Message:  &pbHeader_subchainHeaders,
			}
			suite.Suite.T().Logf("pbHeader: beefymmr: %+v ,subchain headers: %+v", *pbHeader.BeefyMmr, *pbHeader_subchainHeaders.SubchainHeaders)
			// suite.Suite.T().Logf("pbHeader: %#v", pbHeader)

			// mock gp header marshal and unmarshal
			marshalPBHeader, err := pbHeader.Marshal()
			suite.Require().NoError(err)
			suite.Suite.T().Logf("marshalPBHeader: %+v", marshalPBHeader)

			suite.Suite.T().Log("\n\n------------- mock verify on chain -------------")
			// err = gheader.Unmarshal(marshalGHeader)
			var unmarshalPBHeader ibcgptypes.Header
			err = unmarshalPBHeader.Unmarshal(marshalPBHeader)
			suite.Require().NoError(err)
			suite.Suite.T().Logf("unmarshal gHeader: beefymmr: %+v ,subchain headers: %+v", *unmarshalPBHeader.GetBeefyMmr(), unmarshalPBHeader.GetSubchainHeaders().SubchainHeaders)

			// step1:verify signature
			// unmarshalBeefyMmr := unmarshalGPHeader.BeefyMmr
			unmarshalBeefyMmr := pbHeader.BeefyMmr
			// suite.Suite.T().Logf("pbBeefyMMR: %+v", pbBeefyMMR)
			suite.Suite.T().Logf("unmarshal BeefyMmr: %+v", *unmarshalBeefyMmr)

			unmarshalPBSC := unmarshalBeefyMmr.SignedCommitment
			rebuildBSC := ibcgptypes.ToBeefySC(unmarshalPBSC)
			suite.Suite.T().Logf("rebuildBSC: %+v", rebuildBSC)

			suite.Suite.T().Log("\n------------------ VerifySignatures --------------------------")
			err = clientState.VerifySignatures(rebuildBSC, unmarshalBeefyMmr.SignatureProofs)
			suite.Require().NoError(err)
			suite.Suite.T().Log("\n------------------ VerifySignatures end ----------------------\n")

			suite.Suite.T().Log("\n------------------ VerifyMMRBatchProof for beefy  --------------------------")
			// step2, verify mmr
			rebuildMMRLeaves := ibcgptypes.ToBeefyMMRLeaves(unmarshalBeefyMmr.MmrLeavesAndBatchProof.Leaves)
			suite.Suite.T().Logf("rebuildMMRLeaves: %+v", rebuildMMRLeaves)
			rebuildMMRBatchProof := ibcgptypes.ToMMRBatchProof(unmarshalBeefyMmr.MmrLeavesAndBatchProof.MmrBatchProof)
			suite.Suite.T().Logf("Convert2MMRBatchProof: %+v", rebuildMMRBatchProof)
			// check mmr height
			suite.Suite.T().Logf("🐙🐙 ics10::clientState.LatestBeefyHeight: %+v, Commitment.BlockNumber: %+v ", clientState.LatestBeefyHeight.RevisionHeight, unmarshalBeefyMmr.SignedCommitment.Commitment.BlockNumber)
			suite.Require().Less(clientState.LatestBeefyHeight.RevisionHeight, uint64(unmarshalBeefyMmr.SignedCommitment.Commitment.BlockNumber))

			leafIndex := beefy.ConvertBlockNumberToMmrLeafIndex(uint32(beefy.BEEFY_ACTIVATION_BLOCK), bsc.Commitment.BlockNumber)
			calMmrSize := mmr.LeafIndexToMMRSize(uint64(leafIndex))
			suite.Suite.T().Logf("🐙🐙 ics10::CheckHeaderAndUpdateState -> Commitment.BlockNumber: %+v, leafIndex: %+v, cal mmrSize: %+v", bsc.Commitment.BlockNumber, leafIndex, calMmrSize)
			err = clientState.VerifyMMR(rebuildBSC.Commitment.Payload[0].Data, calMmrSize,
				rebuildMMRLeaves, rebuildMMRBatchProof)
			suite.Require().NoError(err)
			// update client height beefy mmr
			clientState.LatestBeefyHeight = clienttypes.NewHeight(clienttypes.ParseChainID(chainID), uint64(latestSignedCommitmentBlockNumber))
			clientState.LatestChainHeight = clienttypes.NewHeight(clienttypes.ParseChainID(chainID), uint64(latestSignedCommitmentBlockNumber))
			clientState.LatestMmrRoot = unmarshalPBSC.Commitment.Payloads[0].Data
			suite.Suite.T().Log("\n------------------ VerifyMMRBatchProof for beefy end ----------------------\n")

			suite.Suite.T().Log("\n------------------mock async VerifyMMRBatchProof for subchain header  --------------------------")
			// step2, async verify mmr for subchain header
			unmarshalSubchainHeadersMmrProof := unmarshalPBHeader.GetSubchainHeaders().MmrLeavesAndBatchProof
			rebuildSubchainHeaderMMRLeaves := ibcgptypes.ToBeefyMMRLeaves(unmarshalSubchainHeadersMmrProof.Leaves)
			suite.Suite.T().Logf("rebuildSubchainHeaderMMRLeaves: %+v", rebuildSubchainHeaderMMRLeaves)
			rebuildSubchainHeaderMMRBatchProof := ibcgptypes.ToMMRBatchProof(unmarshalSubchainHeadersMmrProof.MmrBatchProof)
			suite.Suite.T().Logf("rebuildSubchainHeaderMMRBatchProof: %+v", rebuildSubchainHeaderMMRBatchProof)
			// check : latest block height must less LatestBeefyHeight
			suite.Suite.T().Logf("🐙🐙 ics10::latest block height: %+v, clientState.LatestBeefyHeight: %+v ", unmarshalPBHeader.GetHeight(), clientState.LatestBeefyHeight)
			suite.Require().LessOrEqual(unmarshalPBHeader.GetHeight().GetRevisionHeight(), clientState.LatestBeefyHeight.RevisionHeight)
			// cal mmr size use clientState.LatestBeefyHeight.RevisionHeight
			leafIndex2 := beefy.ConvertBlockNumberToMmrLeafIndex(uint32(beefy.BEEFY_ACTIVATION_BLOCK), uint32(clientState.LatestBeefyHeight.RevisionHeight))
			calMmrSize2 := mmr.LeafIndexToMMRSize(uint64(leafIndex2))
			suite.Suite.T().Logf("🐙🐙 ics10::CheckHeaderAndUpdateState -> clientState.LatestBeefyHeight.RevisionHeight: %+v, leafIndex2: %+v, cal mmrSize: %+v", clientState.LatestBeefyHeight.RevisionHeight, leafIndex2, calMmrSize2)
			err = clientState.VerifyMMR(rebuildBSC.Commitment.Payload[0].Data, calMmrSize,
				rebuildSubchainHeaderMMRLeaves, rebuildSubchainHeaderMMRBatchProof)
			suite.Require().NoError(err)

			suite.Suite.T().Log("\n------------------mock async VerifyMMRBatchProof for subchain header end ----------------------\n")
			// step3, verify header
			// convert pb subchain header to beefy subchain header
			var rebuildSubchainHeaders []beefy.SubchainHeader
			unmarshalSubchainHeaders := unmarshalPBHeader.GetSubchainHeaders().SubchainHeaders
			for _, header := range unmarshalSubchainHeaders {
				rebuildHeader := beefy.SubchainHeader{
					ChainId:     header.ChainId,
					BlockNumber: header.BlockNumber,
					BlockHeader: header.BlockHeader,
					Timestamp:   beefy.StateProof(header.Timestamp),
				}

				rebuildSubchainHeaders = append(rebuildSubchainHeaders, rebuildHeader)
			}

			// suite.Suite.T().Logf("rebuildSolochainHeaderMap: %+v", rebuildSolochainHeaderMap)
			suite.Suite.T().Logf("unmarshal subchainHeaders: %+v", unmarshalSubchainHeaders)

			suite.Suite.T().Log("\n------------------ VerifySubchainHeader --------------------------")

			// cal mmr size use clientState.LatestBeefyHeight.RevisionHeight

			err = clientState.VerifyHeader(unmarshalPBHeader, clientState.LatestMmrRoot, calMmrSize2)
			suite.Require().NoError(err)
			suite.Suite.T().Log("\n------------------ VerifySubchainHeader end ----------------------\n")

			// step4, update client state
			// // update client height
			// clientState.LatestBeefyHeight = clienttypes.NewHeight(clienttypes.ParseChainID(chainID), uint64(latestSignedCommitmentBlockNumber))
			// clientState.LatestChainHeight = clienttypes.NewHeight(clienttypes.ParseChainID(chainID), uint64(latestSignedCommitmentBlockNumber))
			// clientState.LatestMmrRoot = unmarshalPBSC.Commitment.Payloads[0].Data

			// find latest next authority set from mmrleaves
			var latestAuthoritySet *ibcgptypes.BeefyAuthoritySet
			var latestAuthoritySetId uint64
			for _, leaf := range rebuildMMRLeaves {
				if latestAuthoritySetId < uint64(leaf.BeefyNextAuthoritySet.ID) {
					latestAuthoritySetId = uint64(leaf.BeefyNextAuthoritySet.ID)
					latestAuthoritySet = &ibcgptypes.BeefyAuthoritySet{
						Id:   uint64(leaf.BeefyNextAuthoritySet.ID),
						Len:  uint32(leaf.BeefyNextAuthoritySet.Len),
						Root: leaf.BeefyNextAuthoritySet.Root[:],
					}
				}

			}
			suite.Suite.T().Logf("current clientState.AuthoritySet.Id: %+v", clientState.LatestAuthoritySet.Id)
			suite.Suite.T().Logf("latestAuthoritySetId: %+v", latestAuthoritySetId)
			suite.Suite.T().Logf("latestAuthoritySet: %+v", *latestAuthoritySet)

			//  update client state authority set
			if clientState.LatestAuthoritySet.Id < latestAuthoritySetId {
				clientState.LatestAuthoritySet = *latestAuthoritySet
				suite.Suite.T().Logf("update clientState.AuthoritySet : %+v", clientState.LatestAuthoritySet)
			}
			// print latest client state
			suite.Suite.T().Logf("updated client state: %+v", *clientState)

			// step5,update consensue state
			var latestHeight uint32
			for _, header := range unmarshalSubchainHeaders {
				var decodeHeader gsrpctypes.Header
				err = gsrpccodec.Decode(header.BlockHeader, &decodeHeader)
				suite.Require().NoError(err)
				var timestamp gsrpctypes.U64
				err = gsrpccodec.Decode(header.Timestamp.Value, &timestamp)
				suite.Require().NoError(err)

				consensusState := ibcgptypes.ConsensusState{
					Timestamp: time.UnixMilli(int64(timestamp)),
					Root:      decodeHeader.StateRoot[:],
				}
				consensusStateKVStore[uint32(decodeHeader.Number)] = consensusState

				if latestHeight < uint32(decodeHeader.Number) {
					latestHeight = uint32(decodeHeader.Number)
				}
			}
			suite.Suite.T().Logf("latest consensusStateKVStore: %+v", consensusStateKVStore)
			suite.Suite.T().Logf("latest height and consensus state: %d,%+v", latestHeight, consensusStateKVStore[latestHeight])

			// step6, mock to build and verify state proof
			for num, consnesue := range consensusStateKVStore {
				targetBlockHash, err := localSubchainEndpoint.RPC.Chain.GetBlockHash(uint64(num))
				suite.Require().NoError(err)
				timestampProof, err := beefy.GetTimestampProof(localSubchainEndpoint, targetBlockHash)
				suite.Require().NoError(err)

				proofs := make([][]byte, len(timestampProof.Proof))
				for i, proof := range timestampProof.Proof {
					proofs[i] = proof
				}
				suite.Suite.T().Logf("timestampProof proofs: %+v", proofs)
				paraTimestampStoragekey := beefy.CreateStorageKeyPrefix("Timestamp", "Now")
				suite.Suite.T().Logf("timestampStoragekey: %#x", paraTimestampStoragekey)
				timestampValue := consnesue.Timestamp.UnixMilli()
				suite.Suite.T().Logf("timestampValue: %d", timestampValue)
				encodedTimeStampValue, err := gsrpccodec.Encode(timestampValue)
				suite.Require().NoError(err)
				suite.Suite.T().Logf("encodedTimeStampValue: %+v", encodedTimeStampValue)
				suite.Suite.T().Log("\n------------------ VerifyStateProof --------------------------")
				err = beefy.VerifyStateProof(proofs, consnesue.Root, paraTimestampStoragekey, encodedTimeStampValue)
				suite.Require().NoError(err)
				suite.Suite.T().Log("beefy.VerifyStateProof(proof,root,key,value) result: True")
				suite.Suite.T().Log("\n------------------ VerifyStateProof end ----------------------\n")
			}

			received++
			if received >= 3 {
				return
			}
		case <-timeout:
			suite.Suite.T().Logf("timeout reached without getting 2 notifications from subscription")
			return
		}
	}
}

// This test case need to spin up polkadot relay chain and a parachain
func (suite *GrandpaTestSuite) TestParachainLocalNet() {
	suite.Suite.T().Skip("need setup relay chain and parachain")
	localRelayEndpoint, err := gsrpc.NewSubstrateAPI(beefy.LOCAL_RELAY_ENDPPOIT)
	if err != nil {
		suite.Suite.T().Logf("Connecting err: %v", err)
	}
	ch := make(chan interface{})
	sub, err := localRelayEndpoint.Client.Subscribe(
		context.Background(),
		"beefy",
		"subscribeJustifications",
		"unsubscribeJustifications",
		"justifications",
		ch)
	suite.Require().NoError(err)
	suite.Suite.T().Logf("subscribed to relaychain %s\n", beefy.LOCAL_RELAY_ENDPPOIT)
	defer sub.Unsubscribe()

	localParachainEndpoint, err := gsrpc.NewSubstrateAPI(beefy.LOCAL_PARACHAIN_ENDPOINT)
	if err != nil {
		suite.Suite.T().Logf("Connecting err: %v", err)
	}
	suite.Suite.T().Logf("subscribed to parachain %s\n", beefy.LOCAL_PARACHAIN_ENDPOINT)

	timeout := time.After(24 * time.Hour)
	received := 0
	// var preBlockNumber uint32
	// var preBloackHash gsrpctypes.Hash
	var clientState *ibcgptypes.ClientState
	consensusStateKVStore := make(map[uint32]ibcgptypes.ConsensusState)
	for {
		select {
		case msg := <-ch:
			suite.Suite.T().Logf("encoded msg: %s", msg)

			s := &beefy.VersionedFinalityProof{}
			err := gsrpccodec.DecodeFromHex(msg.(string), s)
			if err != nil {
				panic(err)
			}

			suite.Suite.T().Logf("decoded msg: %+v\n", s)
			// suite.Suite.T().Logf("decoded msg: %#v\n", s)
			latestSignedCommitmentBlockNumber := s.SignedCommitment.Commitment.BlockNumber
			// suite.Suite.T().Logf("blockNumber: %d\n", latestBlockNumber)
			latestSignedCommitmentBlockHash, err := localRelayEndpoint.RPC.Chain.GetBlockHash(uint64(latestSignedCommitmentBlockNumber))
			suite.Require().NoError(err)
			suite.Suite.T().Logf("latestSignedCommitmentBlockNumber: %d latestSignedCommitmentBlockHash: %#x",
				latestSignedCommitmentBlockNumber, latestSignedCommitmentBlockHash)

			// build and init grandpa lc client state
			if clientState == nil {
				//  init client state
				suite.Suite.T().Log("clientState == nil, need to init !")
				authoritySet, err := beefy.GetBeefyAuthoritySet(latestSignedCommitmentBlockHash, localRelayEndpoint, "BeefyAuthorities")
				suite.Require().NoError(err)
				suite.Suite.T().Logf("current authority set: %+v", authoritySet)
				nextAuthoritySet, err := beefy.GetBeefyAuthoritySet(latestSignedCommitmentBlockHash, localRelayEndpoint, "BeefyNextAuthorities")
				suite.Require().NoError(err)
				suite.Suite.T().Logf("next authority set: %+v", nextAuthoritySet)
				// get the parachain latest height
				// the height must be inclueded into relay chain, if not exist ,the height is zero
				fromBlockNumber := latestSignedCommitmentBlockNumber - 7 // for test
				suite.Suite.T().Logf("fromBlockNumber: %d toBlockNumber: %d", fromBlockNumber, latestSignedCommitmentBlockNumber)

				fromBlockHash, err := localRelayEndpoint.RPC.Chain.GetBlockHash(uint64(fromBlockNumber))
				suite.Require().NoError(err)
				suite.Suite.T().Logf("fromBlockHash: %#x ", fromBlockHash)
				suite.Suite.T().Logf("toBlockHash: %#x", latestSignedCommitmentBlockHash)
				changeSets, err := beefy.QueryParachainStorage(localRelayEndpoint, beefy.LOCAL_PARACHAIN_ID, fromBlockHash, latestSignedCommitmentBlockHash)
				suite.Require().NoError(err)
				suite.Suite.T().Logf("changeSet len: %d", len(changeSets))
				var latestChainHeight uint32
				if len(changeSets) == 0 {
					latestChainHeight = 0
				} else {
					var packedParachainHeights []int
					for _, changeSet := range changeSets {
						for _, change := range changeSet.Changes {
							suite.Suite.T().Logf("change.StorageKey: %#x", change.StorageKey)
							suite.Suite.T().Log("change.HasStorageData: ", change.HasStorageData)
							suite.Suite.T().Logf("change.HasStorageData: %#x", change.StorageData)
							if change.HasStorageData {
								var parachainheader gsrpctypes.Header
								// first decode byte array
								var bz []byte
								err = gsrpccodec.Decode(change.StorageData, &bz)
								suite.Require().NoError(err)
								// second decode header
								err = gsrpccodec.Decode(bz, &parachainheader)
								suite.Require().NoError(err)
								packedParachainHeights = append(packedParachainHeights, int(parachainheader.Number))
							}
						}

					}
					suite.Suite.T().Logf("raw packedParachainHeights: %+v", packedParachainHeights)
					// sort heights and find latest height
					sort.Sort(sort.Reverse(sort.IntSlice(packedParachainHeights)))
					suite.Suite.T().Logf("sort.Reverse: %+v", packedParachainHeights)
					latestChainHeight = uint32(packedParachainHeights[0])
					suite.Suite.T().Logf("latestHeight: %d", latestChainHeight)
				}

				clientState = &ibcgptypes.ClientState{
					ChainId:     "astar-0",
					ChainType:   beefy.CHAINTYPE_PARACHAIN,
					ParachainId: beefy.LOCAL_PARACHAIN_ID,

					LatestBeefyHeight: clienttypes.NewHeight(clienttypes.ParseChainID("astar-0"), uint64(latestSignedCommitmentBlockNumber)),
					LatestMmrRoot:     s.SignedCommitment.Commitment.Payload[0].Data,
					LatestChainHeight: clienttypes.NewHeight(clienttypes.ParseChainID("astar-0"), uint64(latestChainHeight)),
					FrozenHeight:      clienttypes.NewHeight(clienttypes.ParseChainID("astar-0"), 0),
					LatestAuthoritySet: ibcgptypes.BeefyAuthoritySet{
						Id:   uint64(authoritySet.ID),
						Len:  uint32(authoritySet.Len),
						Root: authoritySet.Root[:],
					},
				}
				// suite.Suite.T().Logf("init client state: %+v", &clientState)
				suite.Suite.T().Logf("init client state: %+v", *clientState)
				// test pb marshal and unmarshal
				marshalCS, err := clientState.Marshal()
				suite.Require().NoError(err)
				suite.Suite.T().Logf("marshal client state: %+v", marshalCS)
				// unmarshal
				// err = clientState.Unmarshal(marshalCS)
				var unmarshalCS ibcgptypes.ClientState
				err = unmarshalCS.Unmarshal(marshalCS)
				suite.Require().NoError(err)
				suite.Suite.T().Logf("unmarshal client state: %+v", unmarshalCS)
				continue
			}

			// step1,build authourities proof for current beefy signatures
			authorities, err := beefy.GetBeefyAuthorities(latestSignedCommitmentBlockHash, localRelayEndpoint, "Authorities")
			suite.Require().NoError(err)
			bsc := beefy.ConvertCommitment(s.SignedCommitment)
			var authorityIdxes []uint64
			for _, v := range bsc.Signatures {
				idx := v.Index
				authorityIdxes = append(authorityIdxes, uint64(idx))
			}
			_, authorityProof, err := beefy.BuildAuthorityProof(authorities, authorityIdxes)
			suite.Require().NoError(err)

			// step2,build beefy mmr
			fromBlockNumber := clientState.LatestBeefyHeight.RevisionHeight + 1

			fromBlockHash, err := localRelayEndpoint.RPC.Chain.GetBlockHash(uint64(fromBlockNumber))
			suite.Require().NoError(err)

			beefyMmrProofHeight := []uint32{uint32(latestSignedCommitmentBlockNumber)}
			// build mmr proofs for leaves containing target paraId
			// mmrBatchProof, err := beefy.BuildMMRBatchProof(localRelayEndpoint, &latestSignedCommitmentBlockHash, targetRelaychainBlockHeights)
			beeMmrProof, err := beefy.BuildMMRProofs(localRelayEndpoint, beefyMmrProofHeight,
				gsrpctypes.NewOptionU32(gsrpctypes.U32(latestSignedCommitmentBlockNumber)),
				gsrpctypes.NewOptionHashEmpty())
			suite.Require().NoError(err)

			pbBeefyMMR := ibcgptypes.ToPBBeefyMMR(bsc, beeMmrProof, authorityProof)
			suite.Suite.T().Logf("pbBeefyMMR: %+v", pbBeefyMMR)

			changeSets, err := beefy.QueryParachainStorage(localRelayEndpoint, beefy.LOCAL_PARACHAIN_ID, fromBlockHash, latestSignedCommitmentBlockHash)
			suite.Require().NoError(err)

			var targetRelaychainBlockHeights []uint32
			for _, changeSet := range changeSets {
				relayHeader, err := localRelayEndpoint.RPC.Chain.GetHeader(changeSet.Block)
				suite.Require().NoError(err)
				targetRelaychainBlockHeights = append(targetRelaychainBlockHeights, uint32(relayHeader.Number))
			}
			parachainMmrProof, err := beefy.BuildMMRProofs(localRelayEndpoint, targetRelaychainBlockHeights,
				gsrpctypes.NewOptionU32(gsrpctypes.U32(latestSignedCommitmentBlockNumber)),
				gsrpctypes.NewOptionHashEmpty())
			suite.Require().NoError(err)
			// step3, build header proof
			// build parachain header proof and verify that proof
			parachainHeaders, err := beefy.BuildParachainHeaders(localRelayEndpoint, localParachainEndpoint,
				parachainMmrProof.Proof.LeafIndexes, "astar-0", beefy.LOCAL_PARACHAIN_ID)
			suite.Require().NoError(err)
			suite.Suite.T().Logf("parachainHeaders: %+v", parachainHeaders)

			// convert beefy parachain header to pb parachain header
			pbHeader_parachains := ibcgptypes.ToPBParachainHeaders(parachainHeaders, parachainMmrProof)
			// suite.Suite.T().Logf("pbHeader_parachains: %+v", pbHeader_parachains)

			// build grandpa pb header
			pbHeader := ibcgptypes.Header{
				BeefyMmr: &pbBeefyMMR,
				Message:  &pbHeader_parachains,
			}
			suite.Suite.T().Logf("pbHeader: beefymmr: %+v ,parachain headers: %+v", *pbHeader.BeefyMmr, *pbHeader_parachains.ParachainHeaders)

			// mock gp header marshal and unmarshal
			marshalPBHeader, err := pbHeader.Marshal()
			suite.Require().NoError(err)
			// suite.Suite.T().Logf("marshal gHeader: %+v", marshalGHeader)

			suite.Suite.T().Log("\n------------- mock verify on chain -------------\n")

			// err = gheader.Unmarshal(marshalGHeader)
			var unmarshalPBHeader ibcgptypes.Header
			err = unmarshalPBHeader.Unmarshal(marshalPBHeader)
			suite.Require().NoError(err)
			suite.Suite.T().Logf("unmarshal gHeader: beefymmr: %+v ,parachain headers: %+v", *unmarshalPBHeader.GetBeefyMmr(), unmarshalPBHeader.GetParachainHeaders().ParachainHeaders)

			// verify signature
			// unmarshalBeefyMmr := unmarshalGPHeader.BeefyMmr
			unmarshalBeefyMmr := pbHeader.BeefyMmr
			// suite.Suite.T().Logf("gBeefyMMR: %+v", gBeefyMMR)
			suite.Suite.T().Logf("unmarshal BeefyMmr: %+v", *unmarshalBeefyMmr)

			unmarshalPBSC := unmarshalBeefyMmr.SignedCommitment

			rebuildBSC := ibcgptypes.ToBeefySC(unmarshalPBSC)
			suite.Suite.T().Logf("rebuildBSC: %+v", rebuildBSC)

			suite.Suite.T().Log("------------------ verify signature ----------------------")
			err = clientState.VerifySignatures(rebuildBSC, unmarshalBeefyMmr.SignatureProofs)
			suite.Require().NoError(err)
			suite.Suite.T().Log("--------------verify signature end ----------------------------")

			// step2, verify mmr
			rebuildMMRLeaves := ibcgptypes.ToBeefyMMRLeaves(unmarshalBeefyMmr.MmrLeavesAndBatchProof.Leaves)
			suite.Suite.T().Logf("rebuildMMRLeaves: %+v", rebuildMMRLeaves)
			beefyMmrBatchProof := ibcgptypes.ToMMRBatchProof(unmarshalBeefyMmr.MmrLeavesAndBatchProof.MmrBatchProof)
			suite.Suite.T().Logf("Convert2MMRBatchProof: %+v", beefyMmrBatchProof)

			suite.Suite.T().Log("\n---------------- verify beefy mmr proof --------------------")
			// check mmr height
			// suite.Require().Less(clientState.LatestBeefyHeight, unmarshalBeefyMmr.SignedCommitment.Commitment.BlockNumber)
			// result, err := beefy.VerifyMMRBatchProof(rebuildBSC.Commitment.Payload,
			// suite.Require().True(result)
			// 	unmarshalBeefyMmr.MmrSize, rebuildMMRLeaves, beefyMmrBatchProof)
			leafIndex := beefy.ConvertBlockNumberToMmrLeafIndex(uint32(beefy.BEEFY_ACTIVATION_BLOCK), bsc.Commitment.BlockNumber)
			calMmrSize := mmr.LeafIndexToMMRSize(uint64(leafIndex))
			suite.Suite.T().Logf("🐙🐙 ics10::CheckHeaderAndUpdateState -> Commitment.BlockNumber: %+v, leafIndex: %+v, cal mmrSize: %+v", bsc.Commitment.BlockNumber, leafIndex, calMmrSize)

			err = clientState.VerifyMMR(rebuildBSC.Commitment.Payload[0].Data, calMmrSize,
				rebuildMMRLeaves, beefyMmrBatchProof)
			suite.Require().NoError(err)

			// update client height beefy mmr
			clientState.LatestBeefyHeight = clienttypes.NewHeight(clienttypes.ParseChainID(chainID), uint64(latestSignedCommitmentBlockNumber))
			// clientState.LatestChainHeight = clienttypes.NewHeight(clienttypes.ParseChainID(chainID), uint64(latestSignedCommitmentBlockNumber))
			clientState.LatestMmrRoot = unmarshalPBSC.Commitment.Payloads[0].Data
			suite.Suite.T().Log("\n-------------verify beefy mmr proof end--------------------\n")

			// step3, verify header
			// convert pb parachain header to beefy parachain header
			var rebuildParachainHeaders []beefy.ParachainHeader
			unmarshalParachainHeaders := unmarshalPBHeader.GetParachainHeaders().ParachainHeaders
			for _, header := range unmarshalParachainHeaders {
				rebuildParachainHeader := beefy.ParachainHeader{
					ChainId:            "astar-0",
					ParaId:             header.ParachainId,
					RelayerChainNumber: header.RelayerChainNumber,
					BlockHeader:        header.BlockHeader,
					Proof:              header.Proofs,
					HeaderIndex:        header.HeaderIndex,
					HeaderCount:        header.HeaderCount,
					Timestamp:          beefy.StateProof(header.Timestamp),
				}
				rebuildParachainHeaders = append(rebuildParachainHeaders, rebuildParachainHeader)
			}
			// suite.Suite.T().Logf("parachainHeaderMap: %+v", parachainHeaderMap)
			// suite.Suite.T().Logf("gParachainHeaderMap: %+v", gParachainHeaderMap)
			suite.Suite.T().Logf("unmarshal parachainHeaders: %+v", unmarshalParachainHeaders)
			// suite.Suite.T().Logf("rebuildSolochainHeaderMap: %+v", rebuildParachainHeaderMap)
			suite.Suite.T().Log("\n----------- VerifyParachainHeader -----------")
			// err = beefy.VerifyParachainHeader(rebuildMMRLeaves, rebuildParachainHeaderMap)
			// suite.Require().NoError(err)
			err = clientState.VerifyHeader(unmarshalPBHeader, clientState.LatestMmrRoot, calMmrSize)
			suite.Require().NoError(err)
			suite.Suite.T().Log("\n--------------------------------------------\n")

			// step4, update client state
			clientState.LatestBeefyHeight = clienttypes.NewHeight(clienttypes.ParseChainID(chainID), uint64(latestSignedCommitmentBlockNumber))
			// clientState.LatestChainHeight = latestSignedCommitmentBlockNumber
			clientState.LatestMmrRoot = unmarshalPBSC.Commitment.Payloads[0].Data
			// find latest next authority set from mmrleaves
			var latestAuthoritySet *ibcgptypes.BeefyAuthoritySet
			var latestAuthoritySetId uint64
			for _, leaf := range rebuildMMRLeaves {
				if latestAuthoritySetId < uint64(leaf.BeefyNextAuthoritySet.ID) {
					latestAuthoritySetId = uint64(leaf.BeefyNextAuthoritySet.ID)
					latestAuthoritySet = &ibcgptypes.BeefyAuthoritySet{
						Id:   uint64(leaf.BeefyNextAuthoritySet.ID),
						Len:  uint32(leaf.BeefyNextAuthoritySet.Len),
						Root: leaf.BeefyNextAuthoritySet.Root[:],
					}
				}

			}
			suite.Suite.T().Logf("current clientState.AuthoritySet.Id: %+v", clientState.LatestAuthoritySet.Id)
			// suite.Suite.T().Logf("latestAuthoritySetId: %+v", latestAuthoritySetId)
			suite.Suite.T().Logf("latestAuthoritySet: %+v", *latestAuthoritySet)
			//  update client state authority set
			if clientState.LatestAuthoritySet.Id < latestAuthoritySetId {
				clientState.LatestAuthoritySet = *latestAuthoritySet
				suite.Suite.T().Logf("update clientState.AuthoritySet : %+v", clientState.LatestAuthoritySet)
			}

			// print latest client state
			suite.Suite.T().Logf("updated client state: %+v", *clientState)

			// step5,update consensue state
			var latestHeight uint32
			for _, header := range unmarshalParachainHeaders {
				var decodeHeader gsrpctypes.Header
				err = gsrpccodec.Decode(header.BlockHeader, &decodeHeader)
				suite.Require().NoError(err)
				var timestamp gsrpctypes.U64
				err = gsrpccodec.Decode(header.Timestamp.Value, &timestamp)
				suite.Require().NoError(err)

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
			suite.Suite.T().Logf("latest consensusStateKVStore: %+v", consensusStateKVStore)
			suite.Suite.T().Logf("latest height and consensus state: %d,%+v", latestHeight, consensusStateKVStore[latestHeight])

			// update client state latest chain height
			clientState.LatestChainHeight = clienttypes.NewHeight(clienttypes.ParseChainID(clientState.ChainId), uint64(latestHeight))

			// step6, mock to build and verify state proof
			for num, consnesue := range consensusStateKVStore {
				// Note: get data from parachain
				targetBlockHash, err := localParachainEndpoint.RPC.Chain.GetBlockHash(uint64(num))
				suite.Require().NoError(err)
				timestampProof, err := beefy.GetTimestampProof(localParachainEndpoint, targetBlockHash)
				suite.Require().NoError(err)

				proofs := make([][]byte, len(timestampProof.Proof))
				for i, proof := range timestampProof.Proof {
					proofs[i] = proof
				}
				suite.Suite.T().Logf("timestampProof proofs: %+v", proofs)
				paraTimestampStoragekey := beefy.CreateStorageKeyPrefix("Timestamp", "Now")
				suite.Suite.T().Logf("paraTimestampStoragekey: %#x", paraTimestampStoragekey)
				timestampValue := consnesue.Timestamp.UnixMilli()
				suite.Suite.T().Logf("timestampValue: %d", timestampValue)
				encodedTimeStampValue, err := gsrpccodec.Encode(timestampValue)
				suite.Require().NoError(err)
				suite.Suite.T().Logf("encodedTimeStampValue: %+v", encodedTimeStampValue)
				suite.Suite.T().Log("\n------------- mock verify state proof-------------")

				err = beefy.VerifyStateProof(proofs, consnesue.Root, paraTimestampStoragekey, encodedTimeStampValue)
				suite.Require().NoError(err)
				suite.Suite.T().Log("beefy.VerifyStateProof(proof,root,key,value) result: True")
				suite.Suite.T().Log("\n--------------------------------------------------")
			}

			received++
			if received >= 2 {
				return
			}
		case <-timeout:
			suite.Suite.T().Logf("timeout reached without getting 2 notifications from subscription")
			return
		}
	}
}

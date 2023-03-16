package types_test

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/octopus-network/beefy-go/beefy"
	tmtypes "github.com/tendermint/tendermint/types"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	gsrpctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	gsrpccodec "github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	clienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v5/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"
	types "github.com/cosmos/ibc-go/v5/modules/light-clients/07-tendermint/types"
	ibcgptypes "github.com/cosmos/ibc-go/v5/modules/light-clients/10-grandpa/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
	ibctestingmock "github.com/cosmos/ibc-go/v5/testing/mock"
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

func (suite *GrandpaTestSuite) TestPruneConsensusState() {
	// create path and setup clients
	path := ibctesting.NewPath(suite.chainA, suite.chainB)
	suite.coordinator.SetupClients(path)

	// get the first height as it will be pruned first.
	var pruneHeight exported.Height
	getFirstHeightCb := func(height exported.Height) bool {
		pruneHeight = height
		return true
	}
	ctx := path.EndpointA.Chain.GetContext()
	clientStore := path.EndpointA.Chain.App.GetIBCKeeper().ClientKeeper.ClientStore(ctx, path.EndpointA.ClientID)
	err := types.IterateConsensusStateAscending(clientStore, getFirstHeightCb)
	suite.Require().Nil(err)

	// this height will be expired but not pruned
	path.EndpointA.UpdateClient()
	expiredHeight := path.EndpointA.GetClientState().GetLatestHeight()

	// expected values that must still remain in store after pruning
	expectedConsState, ok := path.EndpointA.Chain.GetConsensusState(path.EndpointA.ClientID, expiredHeight)
	suite.Require().True(ok)
	ctx = path.EndpointA.Chain.GetContext()
	clientStore = path.EndpointA.Chain.App.GetIBCKeeper().ClientKeeper.ClientStore(ctx, path.EndpointA.ClientID)
	expectedProcessTime, ok := types.GetProcessedTime(clientStore, expiredHeight)
	suite.Require().True(ok)
	expectedProcessHeight, ok := types.GetProcessedHeight(clientStore, expiredHeight)
	suite.Require().True(ok)
	expectedConsKey := types.GetIterationKey(clientStore, expiredHeight)
	suite.Require().NotNil(expectedConsKey)

	// Increment the time by a week
	suite.coordinator.IncrementTimeBy(7 * 24 * time.Hour)

	// create the consensus state that can be used as trusted height for next update
	path.EndpointA.UpdateClient()

	// Increment the time by another week, then update the client.
	// This will cause the first two consensus states to become expired.
	suite.coordinator.IncrementTimeBy(7 * 24 * time.Hour)
	path.EndpointA.UpdateClient()

	ctx = path.EndpointA.Chain.GetContext()
	clientStore = path.EndpointA.Chain.App.GetIBCKeeper().ClientKeeper.ClientStore(ctx, path.EndpointA.ClientID)

	// check that the first expired consensus state got deleted along with all associated metadata
	consState, ok := path.EndpointA.Chain.GetConsensusState(path.EndpointA.ClientID, pruneHeight)
	suite.Require().Nil(consState, "expired consensus state not pruned")
	suite.Require().False(ok)
	// check processed time metadata is pruned
	processTime, ok := types.GetProcessedTime(clientStore, pruneHeight)
	suite.Require().Equal(uint64(0), processTime, "processed time metadata not pruned")
	suite.Require().False(ok)
	processHeight, ok := types.GetProcessedHeight(clientStore, pruneHeight)
	suite.Require().Nil(processHeight, "processed height metadata not pruned")
	suite.Require().False(ok)

	// check iteration key metadata is pruned
	consKey := types.GetIterationKey(clientStore, pruneHeight)
	suite.Require().Nil(consKey, "iteration key not pruned")

	// check that second expired consensus state doesn't get deleted
	// this ensures that there is a cap on gas cost of UpdateClient
	consState, ok = path.EndpointA.Chain.GetConsensusState(path.EndpointA.ClientID, expiredHeight)
	suite.Require().Equal(expectedConsState, consState, "consensus state incorrectly pruned")
	suite.Require().True(ok)
	// check processed time metadata is not pruned
	processTime, ok = types.GetProcessedTime(clientStore, expiredHeight)
	suite.Require().Equal(expectedProcessTime, processTime, "processed time metadata incorrectly pruned")
	suite.Require().True(ok)

	// check processed height metadata is not pruned
	processHeight, ok = types.GetProcessedHeight(clientStore, expiredHeight)
	suite.Require().Equal(expectedProcessHeight, processHeight, "processed height metadata incorrectly pruned")
	suite.Require().True(ok)

	// check iteration key metadata is not pruned
	consKey = types.GetIterationKey(clientStore, expiredHeight)
	suite.Require().Equal(expectedConsKey, consKey, "iteration key incorrectly pruned")
}

func (suite *GrandpaTestSuite) TestSolochainLocalNet() {
	suite.Suite.T().Skip("need to setup relaychain or solochain")
	localSolochainEndpoint, err := gsrpc.NewSubstrateAPI(beefy.LOCAL_RELAY_ENDPPOIT)
	if err != nil {
		suite.Suite.T().Logf("Connecting err: %v", err)
	}
	ch := make(chan interface{})
	sub, err := localSolochainEndpoint.Client.Subscribe(
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
			latestSignedCommitmentBlockHash, err := localSolochainEndpoint.RPC.Chain.GetBlockHash(uint64(latestSignedCommitmentBlockNumber))
			suite.Require().NoError(err)
			suite.Suite.T().Logf("latestSignedCommitmentBlockNumber: %d latestSignedCommitmentBlockHash: %#x",
				latestSignedCommitmentBlockNumber, latestSignedCommitmentBlockHash)

			// build and init grandpa lc client state
			if clientState == nil {
				//  init client state
				suite.Suite.T().Log("clientState == nil, need to init !")
				authoritySet, err := beefy.GetBeefyAuthoritySet(latestSignedCommitmentBlockHash, localSolochainEndpoint, "BeefyAuthorities")
				suite.Require().NoError(err)
				suite.Suite.T().Logf("current authority set: %+v", authoritySet)
				nextAuthoritySet, err := beefy.GetBeefyAuthoritySet(latestSignedCommitmentBlockHash, localSolochainEndpoint, "BeefyNextAuthorities")
				suite.Require().NoError(err)
				suite.Suite.T().Logf("next authority set: %+v", nextAuthoritySet)
				// get the parachain or solochain latest height
				// the height must be inclueded into relay chain, if not exist ,the height is zero
				fromBlockNumber := latestSignedCommitmentBlockNumber - 7 // for test
				suite.Suite.T().Logf("fromBlockNumber: %d toBlockNumber: %d", fromBlockNumber, latestSignedCommitmentBlockNumber)

				fromBlockHash, err := localSolochainEndpoint.RPC.Chain.GetBlockHash(uint64(fromBlockNumber))
				suite.Require().NoError(err)
				suite.Suite.T().Logf("fromBlockHash: %#x ", fromBlockHash)
				suite.Suite.T().Logf("toBlockHash: %#x", latestSignedCommitmentBlockHash)

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
				suite.Suite.T().Logf("init client state: %+v", &clientState)
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
			authorities, err := beefy.GetBeefyAuthorities(latestSignedCommitmentBlockHash, localSolochainEndpoint, "Authorities")
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
			targetHeights := []uint32{uint32(latestSignedCommitmentBlockNumber - 1)}
			// build mmr proofs for leaves containing target paraId
			// mmrBatchProof, err := beefy.BuildMMRBatchProof(localSolochainEndpoint, &latestSignedCommitmentBlockHash, targetHeights)
			mmrBatchProof, err := beefy.BuildMMRProofs(localSolochainEndpoint, targetHeights,
				gsrpctypes.NewOptionU32(gsrpctypes.U32(latestSignedCommitmentBlockNumber)),
				gsrpctypes.NewOptionHashEmpty())
			suite.Require().NoError(err)
			pbBeefyMMR := ibcgptypes.ToPBBeefyMMR(bsc, mmrBatchProof, authorityProof)
			suite.Suite.T().Logf("pbBeefyMMR: %+v", pbBeefyMMR)

			// step3, build header proof
			// build solochain header map
			solochainHeaderMap, err := beefy.BuildSolochainHeaderMap(localSolochainEndpoint, mmrBatchProof.Proof.LeafIndexes)
			suite.Require().NoError(err)
			suite.Suite.T().Logf("solochainHeaderMap: %+v", solochainHeaderMap)
			suite.Suite.T().Logf("solochainHeaderMap: %#v", solochainHeaderMap)

			pbHeader_solochainMap := ibcgptypes.ToPBSolochainHeaderMap(solochainHeaderMap)
			// build grandpa pb header
			pbHeader := ibcgptypes.Header{
				BeefyMmr: pbBeefyMMR,
				Message:  &pbHeader_solochainMap,
			}
			suite.Suite.T().Logf("pbHeader: %+v", pbHeader)
			suite.Suite.T().Logf("pbHeader: %#v", pbHeader)

			// mock gp header marshal and unmarshal
			marshalPBHeader, err := pbHeader.Marshal()
			suite.Require().NoError(err)
			suite.Suite.T().Logf("marshalPBHeader: %+v", marshalPBHeader)

			suite.Suite.T().Log("\n\n------------- mock verify on chain -------------")
			// err = gheader.Unmarshal(marshalGHeader)
			var unmarshalPBHeader ibcgptypes.Header
			err = unmarshalPBHeader.Unmarshal(marshalPBHeader)
			suite.Require().NoError(err)
			suite.Suite.T().Logf("unmarshal gHeader: %+v", unmarshalPBHeader)

			// step1:verify signature
			// unmarshalBeefyMmr := unmarshalGPHeader.BeefyMmr
			unmarshalBeefyMmr := pbHeader.BeefyMmr
			// suite.Suite.T().Logf("pbBeefyMMR: %+v", pbBeefyMMR)
			suite.Suite.T().Logf("unmarshal BeefyMmr: %+v", unmarshalBeefyMmr)

			unmarshalPBSC := unmarshalBeefyMmr.SignedCommitment
			rebuildBSC := ibcgptypes.ToBeefySC(unmarshalPBSC)
			suite.Suite.T().Logf("rebuildBSC: %+v", rebuildBSC)

			suite.Suite.T().Log("\n------------------ VerifySignatures --------------------------")
			err = clientState.VerifySignatures(rebuildBSC, unmarshalBeefyMmr.SignatureProofs)
			suite.Require().NoError(err)
			suite.Suite.T().Log("\n------------------ VerifySignatures end ----------------------\n")

			// step2, verify mmr
			rebuildMMRLeaves := ibcgptypes.ToBeefyMMRLeaves(unmarshalBeefyMmr.MmrLeavesAndBatchProof.Leaves)
			suite.Suite.T().Logf("rebuildMMRLeaves: %+v", rebuildMMRLeaves)
			rebuildMMRBatchProof := ibcgptypes.ToMMRBatchProof(unmarshalBeefyMmr.MmrLeavesAndBatchProof)
			suite.Suite.T().Logf("Convert2MMRBatchProof: %+v", rebuildMMRBatchProof)

			suite.Suite.T().Log("\n------------------ VerifyMMRBatchProof --------------------------")
			// check mmr height
			// suite.Require().Less(clientState.LatestBeefyHeight, unmarshalBeefyMmr.SignedCommitment.Commitment.BlockNumber)
			// result, err := beefy.VerifyMMRBatchProof(rebuildBSC.Commitment.Payload,
			// 	unmarshalBeefyMmr.MmrSize, rebuildMMRLeaves, rebuildMMRBatchProof)
			// 	suite.Require().NoError(err)
			// 	suite.Require().True(result)
			err = clientState.VerifyMMR(rebuildBSC, unmarshalBeefyMmr.MmrSize,
				rebuildMMRLeaves, rebuildMMRBatchProof)
			suite.Require().NoError(err)
			suite.Suite.T().Log("\n------------------ VerifyMMRBatchProof end ----------------------\n")

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
			// suite.Suite.T().Logf("rebuildSolochainHeaderMap: %+v", rebuildSolochainHeaderMap)
			suite.Suite.T().Logf("unmarshal solochainHeaderMap: %+v", *unmarshalSolochainHeaderMap)

			suite.Suite.T().Log("\n------------------ VerifySolochainHeader --------------------------")
			// err = beefy.VerifySolochainHeader(rebuildMMRLeaves, rebuildSolochainHeaderMap)
			// suite.Require().NoError(err)
			err = clientState.VerifyHeader(unmarshalPBHeader, rebuildMMRLeaves)
			suite.Require().NoError(err)
			suite.Suite.T().Log("\n------------------ VerifySolochainHeader end ----------------------\n")

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
			suite.Suite.T().Logf("current clientState.AuthoritySet.Id: %+v", clientState.AuthoritySet.Id)
			suite.Suite.T().Logf("latestAuthoritySetId: %+v", latestAuthoritySetId)
			suite.Suite.T().Logf("latestNextAuthoritySet: %+v", *latestNextAuthoritySet)
			//  update client state authority set
			if clientState.AuthoritySet.Id < latestAuthoritySetId {
				clientState.AuthoritySet = *latestNextAuthoritySet
				suite.Suite.T().Logf("update clientState.AuthoritySet : %+v", clientState.AuthoritySet)
			}
			// print latest client state
			suite.Suite.T().Logf("updated client state: %+v", clientState)

			// step5,update consensue state
			var latestHeight uint32
			for _, header := range unmarshalSolochainHeaderMap.SolochainHeaderMap {
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
				targetBlockHash, err := localSolochainEndpoint.RPC.Chain.GetBlockHash(uint64(num))
				suite.Require().NoError(err)
				timestampProof, err := beefy.GetTimestampProof(localSolochainEndpoint, targetBlockHash)
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
			if received >= 1 {
				return
			}
		case <-timeout:
			suite.Suite.T().Logf("timeout reached without getting 2 notifications from subscription")
			return
		}
	}
}

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
	suite.Suite.T().Logf("subscribed to parachain %s\n", beefy.LOCAL_RELAY_ENDPPOIT)

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
			suite.Suite.T().Logf("decoded msg: %#v\n", s)
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
				suite.Suite.T().Logf("init client state: %+v", &clientState)
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
			fromBlockNumber := clientState.LatestBeefyHeight + 1

			fromBlockHash, err := localRelayEndpoint.RPC.Chain.GetBlockHash(uint64(fromBlockNumber))
			suite.Require().NoError(err)

			changeSets, err := beefy.QueryParachainStorage(localRelayEndpoint, beefy.LOCAL_PARACHAIN_ID, fromBlockHash, latestSignedCommitmentBlockHash)
			suite.Require().NoError(err)

			var targetRelaychainBlockHeights []uint32
			for _, changeSet := range changeSets {
				relayHeader, err := localRelayEndpoint.RPC.Chain.GetHeader(changeSet.Block)
				suite.Require().NoError(err)
				targetRelaychainBlockHeights = append(targetRelaychainBlockHeights, uint32(relayHeader.Number))
			}

			// build mmr proofs for leaves containing target paraId
			// mmrBatchProof, err := beefy.BuildMMRBatchProof(localRelayEndpoint, &latestSignedCommitmentBlockHash, targetRelaychainBlockHeights)
			mmrBatchProof, err := beefy.BuildMMRProofs(localRelayEndpoint, targetRelaychainBlockHeights,
				gsrpctypes.NewOptionU32(gsrpctypes.U32(latestSignedCommitmentBlockNumber)),
				gsrpctypes.NewOptionHashEmpty())
			suite.Require().NoError(err)

			pbBeefyMMR := ibcgptypes.ToPBBeefyMMR(bsc, mmrBatchProof, authorityProof)
			suite.Suite.T().Logf("pbBeefyMMR: %+v", pbBeefyMMR)

			// step3, build header proof
			// build parachain header proof and verify that proof
			parachainHeaderMap, err := beefy.BuildParachainHeaderMap(localRelayEndpoint, localParachainEndpoint,
				mmrBatchProof.Proof.LeafIndexes, beefy.LOCAL_PARACHAIN_ID)
			suite.Require().NoError(err)
			suite.Suite.T().Logf("parachainHeaderMap: %+v", parachainHeaderMap)

			// convert beefy parachain header to pb parachain header
			pbHeader_parachainMap := ibcgptypes.ToPBParachainHeaderMap(parachainHeaderMap)
			suite.Suite.T().Logf("pbHeader_parachainMap: %+v", pbHeader_parachainMap)

			// build grandpa pb header
			pbHeader := ibcgptypes.Header{
				BeefyMmr: pbBeefyMMR,
				Message:  &pbHeader_parachainMap,
			}
			suite.Suite.T().Logf("gpheader: %+v", pbHeader)

			// mock gp header marshal and unmarshal
			marshalPBHeader, err := pbHeader.Marshal()
			suite.Require().NoError(err)
			// suite.Suite.T().Logf("marshal gHeader: %+v", marshalGHeader)

			suite.Suite.T().Log("\n------------- mock verify on chain -------------\n")

			// err = gheader.Unmarshal(marshalGHeader)
			var unmarshalPBHeader ibcgptypes.Header
			err = unmarshalPBHeader.Unmarshal(marshalPBHeader)
			suite.Require().NoError(err)
			suite.Suite.T().Logf("unmarshal gHeader: %+v", unmarshalPBHeader)

			// verify signature
			// unmarshalBeefyMmr := unmarshalGPHeader.BeefyMmr
			unmarshalBeefyMmr := pbHeader.BeefyMmr
			// suite.Suite.T().Logf("gBeefyMMR: %+v", gBeefyMMR)
			suite.Suite.T().Logf("unmarshal BeefyMmr: %+v", unmarshalBeefyMmr)

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
			beefyMmrBatchProof := ibcgptypes.ToMMRBatchProof(unmarshalBeefyMmr.MmrLeavesAndBatchProof)
			suite.Suite.T().Logf("Convert2MMRBatchProof: %+v", beefyMmrBatchProof)

			suite.Suite.T().Log("\n---------------- verify mmr proof --------------------")
			// check mmr height
			// suite.Require().Less(clientState.LatestBeefyHeight, unmarshalBeefyMmr.SignedCommitment.Commitment.BlockNumber)
			// result, err := beefy.VerifyMMRBatchProof(rebuildBSC.Commitment.Payload,
			// suite.Require().True(result)
			// 	unmarshalBeefyMmr.MmrSize, rebuildMMRLeaves, beefyMmrBatchProof)
			err = clientState.VerifyMMR(rebuildBSC, unmarshalBeefyMmr.MmrSize,
				rebuildMMRLeaves, beefyMmrBatchProof)
			suite.Require().NoError(err)
			suite.Suite.T().Log("\n-------------verify mmr proof end--------------------\n")

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
			// suite.Suite.T().Logf("parachainHeaderMap: %+v", parachainHeaderMap)
			// suite.Suite.T().Logf("gParachainHeaderMap: %+v", gParachainHeaderMap)
			suite.Suite.T().Logf("unmarshal parachainHeaderMap: %+v", *unmarshalParachainHeaderMap)
			// suite.Suite.T().Logf("rebuildSolochainHeaderMap: %+v", rebuildParachainHeaderMap)
			suite.Suite.T().Log("\n----------- VerifyParachainHeader -----------")
			// err = beefy.VerifyParachainHeader(rebuildMMRLeaves, rebuildParachainHeaderMap)
			// suite.Require().NoError(err)
			err = clientState.VerifyHeader(unmarshalPBHeader, rebuildMMRLeaves)
			suite.Require().NoError(err)
			suite.Suite.T().Log("\n--------------------------------------------\n")

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
			suite.Suite.T().Logf("current clientState.AuthoritySet.Id: %+v", clientState.AuthoritySet.Id)
			suite.Suite.T().Logf("latestAuthoritySetId: %+v", latestAuthoritySetId)
			suite.Suite.T().Logf("latestNextAuthoritySet: %+v", *latestNextAuthoritySet)
			//  update client state authority set
			if clientState.AuthoritySet.Id < latestAuthoritySetId {
				clientState.AuthoritySet = *latestNextAuthoritySet
				suite.Suite.T().Logf("update clientState.AuthoritySet : %+v", clientState.AuthoritySet)
			}

			// print latest client state
			suite.Suite.T().Logf("updated client state: %+v", clientState)

			// step5,update consensue state
			var latestHeight uint32
			for _, header := range unmarshalParachainHeaderMap.ParachainHeaderMap {
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
			if received >= 3 {
				return
			}
		case <-timeout:
			suite.Suite.T().Logf("timeout reached without getting 2 notifications from subscription")
			return
		}
	}
}

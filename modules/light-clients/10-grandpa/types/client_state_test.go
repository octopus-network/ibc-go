package types_test

import (
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	"github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	ibcgptypes "github.com/cosmos/ibc-go/v6/modules/light-clients/10-grandpa/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
)

const (
	testClientID     = "clientidone"
	testConnectionID = "connectionid"
	testPortID       = "testportid"
	testChannelID    = "testchannelid"
	testSequence     = 1

	// Do not change the length of these variables
	fiftyCharChainID    = "12345678901234567890123456789012345678901234567890"
	fiftyOneCharChainID = "123456789012345678901234567890123456789012345678901"
)

var invalidProof = []byte("invalid proof")

func (suite *GrandpaTestSuite) TestStatus() {
	var (
		path        *ibctesting.Path
		clientState *types.ClientState
	)

	testCases := []struct {
		name      string
		malleate  func()
		expStatus exported.Status
	}{
		{"client is active", func() {}, exported.Active},
		{"client is frozen", func() {
			clientState.FrozenHeight = clienttypes.NewHeight(0, 1)
			path.EndpointA.SetClientState(clientState)
		}, exported.Frozen},
		{"client status without consensus state", func() {
			clientState.LatestHeight = clientState.LatestHeight.Increment().(clienttypes.Height)
			path.EndpointA.SetClientState(clientState)
		}, exported.Expired},
		{"client status is expired", func() {
			suite.coordinator.IncrementTimeBy(clientState.TrustingPeriod)
		}, exported.Expired},
	}

	for _, tc := range testCases {
		path = ibctesting.NewPath(suite.chainA, suite.chainB)
		suite.coordinator.SetupClients(path)

		clientStore := suite.chainA.App.GetIBCKeeper().ClientKeeper.ClientStore(suite.chainA.GetContext(), path.EndpointA.ClientID)
		clientState = path.EndpointA.GetClientState().(*types.ClientState)

		tc.malleate()

		status := clientState.Status(suite.chainA.GetContext(), clientStore, suite.chainA.App.AppCodec())
		suite.Require().Equal(tc.expStatus, status)

	}

}

func (suite *GrandpaTestSuite) TestValidate() {
	var clientState *ibcgptypes.ClientState
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{"valid client",
			func() {
				clientState = &gpClientState
			}, true,
		},
		{"beefy height must > zero",
			func() {
				gpClientState.LatestBeefyHeight = clienttypes.NewHeight(clienttypes.ParseChainID(chainID), 0)
				clientState = &gpClientState
			}, false,
		},
	}

	for i, tc := range testCases {
		tc.malleate()
		err := clientState.Validate()
		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			suite.Suite.T().Logf("clientState.Validate() err: %+v", err)
		}
	}
}

func (suite *GrandpaTestSuite) TestInitialize() {
	suite.T().Skip("")
}

func (suite *GrandpaTestSuite) TestVerifyClientConsensusState() {
	suite.T().Skip("")

}

// test verification of the connection on chainB being represented in the
// light client on chainA
func (suite *GrandpaTestSuite) TestVerifyConnectionState() {
	suite.T().Skip("")

}

// test verification of the channel on chainB being represented in the light
// client on chainA
func (suite *GrandpaTestSuite) TestVerifyChannelState() {
	suite.T().Skip("")

}

// test verification of the packet commitment on chainB being represented
// in the light client on chainA. A send from chainB to chainA is simulated.
func (suite *GrandpaTestSuite) TestVerifyPacketCommitment() {
	suite.T().Skip("")
}

// test verification of the acknowledgement on chainB being represented
// in the light client on chainA. A send and ack from chainA to chainB
// is simulated.
func (suite *GrandpaTestSuite) TestVerifyPacketAcknowledgement() {
	suite.T().Skip("")

}

// test verification of the absent acknowledgement on chainB being represented
// in the light client on chainA. A send from chainB to chainA is simulated, but
// no receive.
func (suite *GrandpaTestSuite) TestVerifyPacketReceiptAbsence() {
	suite.T().Skip("")
}

// test verification of the next receive sequence on chainB being represented
// in the light client on chainA. A send and receive from chainB to chainA is
// simulated.
func (suite *GrandpaTestSuite) TestVerifyNextSeqRecv() {
	suite.T().Skip("")

}

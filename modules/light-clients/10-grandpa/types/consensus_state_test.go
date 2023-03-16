package types_test

import (
	"time"

	commitmenttypes "github.com/cosmos/ibc-go/v5/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"
	ibcgptypes "github.com/cosmos/ibc-go/v5/modules/light-clients/10-grandpa/types"
)

func (suite *GrandpaTestSuite) TestConsensusStateValidateBasic() {
	testCases := []struct {
		msg            string
		consensusState *ibcgptypes.ConsensusState
		expectPass     bool
	}{
		{
			"success",
			&consensusState,
			true,
		},

		{
			"root is nil",
			&ibcgptypes.ConsensusState{
				Timestamp: consensusState.Timestamp,
				Root:      []byte{},
			},
			false,
		},

		{
			"timestamp is zero",
			&ibcgptypes.ConsensusState{
				Timestamp: time.Time{},
				Root:      consensusState.Root,
			},
			false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		// check just to increase coverage
		suite.Require().Equal(exported.Grandpa, tc.consensusState.ClientType())

		// Note: consensusState.GetRoot() != consensusState.Root, it`s different type
		suite.Require().NotEqual(tc.consensusState.GetRoot(), tc.consensusState.Root)
		suite.Require().Equal(tc.consensusState.GetRoot(), commitmenttypes.NewMerkleRoot([]byte(tc.consensusState.Root)))

		err := tc.consensusState.ValidateBasic()
		if tc.expectPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.msg)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.msg)
		}
	}
}

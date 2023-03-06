package types

import (
	"reflect"
	time "time"

	gsrpctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	gsrpccodec "github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v3/modules/core/exported"
	"github.com/octopus-network/beefy-go/beefy"
)

var _ exported.Header = &Header{}

// ConsensusState returns the updated consensus state associated with the header
func (h Header) ConsensusState() *ConsensusState {

	//get latest header and time
	latestHeader, latestTime, err := getLastestBlockHeader(h)
	if err != nil {
		Logger.Error("LightClient:", "10-Grandpa", "method:", "getLastestHeader error: ", err)
		return nil
	}
	// build consensue state
	return &ConsensusState{
		Root:      latestHeader.StateRoot[:],
		Timestamp: latestTime,
	}
}

// ClientType defines that the Header is a grandpa consensus algorithm
func (h Header) ClientType() string {
	return exported.Grandpa
}

// GetHeight returns the current height. It returns 0 if the grandpa
// header is nil.
// NOTE: the header.Header is checked to be non nil in ValidateBasic.
func (h Header) GetHeight() exported.Height {
	// if err := h.ValidateBasic(); err != nil {
	// 	return clienttypes.NewHeight(0, 0)
	// }

	// use the beefy height as header height
	latestBeefyHeight := h.BeefyMmr.SignedCommitment.Commitment.BlockNumber
	return clienttypes.NewHeight(0, uint64(latestBeefyHeight))
	// revision := clienttypes.ParseChainID(h.Header.ChainID)
	// return clienttypes.NewHeight(revision, uint64(h.Header.Height))
}

// // GetTime returns the current block timestamp. It returns a zero time if
// // the tendermint header is nil.
// // NOTE: the header.Header is checked to be non nil in ValidateBasic.
func (h Header) GetTime() time.Time {
	// if err := h.ValidateBasic(); err != nil {
	// 	return time.Unix(0, 0)
	// }
	//get latest header and time
	_, latestTime, err := getLastestBlockHeader(h)
	if err != nil {
		Logger.Error("LightClient:", "10-Grandpa", "method:", "getLastestHeader error: ", err)
		return time.Unix(0, 0)
	}
	return latestTime

}

func getLastestBlockHeader(h Header) (gsrpctypes.Header, time.Time, error) {
	var latestHeader *gsrpctypes.Header
	var latestTimestamp time.Time
	var latestHeight uint32
	headerMessage := h.GetMessage()
	switch headerMap := headerMessage.(type) {
	case *Header_SolochainHeaderMap:
		solochainHeaderMap := headerMap.SolochainHeaderMap.SolochainHeaderMap
		for num := range solochainHeaderMap {
			if latestHeight < num {
				latestHeight = num
			}
		}
		solochainHeader := solochainHeaderMap[latestHeight]
		var decodeHeader gsrpctypes.Header
		err := gsrpccodec.Decode(solochainHeader.BlockHeader, &decodeHeader)
		if err != nil {
			return *latestHeader, latestTimestamp, sdkerrors.Wrapf(err, "decode header error")
		}

		// verify timestamp and get it
		err = beefy.VerifyStateProof(solochainHeader.Timestamp.Proofs,
			decodeHeader.StateRoot[:], solochainHeader.Timestamp.Key,
			solochainHeader.Timestamp.Value)
		if err != nil {
			Logger.Error("LightClient:", "10-Grandpa", "method:", "VerifyStateProof error: ", err)
			return *latestHeader, latestTimestamp, sdkerrors.Wrapf(err, "verify timestamp error")
		}
		//decode
		var decodeTimestamp gsrpctypes.U64
		err = gsrpccodec.Decode(solochainHeader.Timestamp.Value, &decodeTimestamp)
		if err != nil {
			return *latestHeader, latestTimestamp, sdkerrors.Wrapf(err, "decode timestamp error")
		}
		latestTimestamp = time.UnixMilli(int64(decodeTimestamp))

	case *Header_ParachainHeaderMap:
		parachainHeaderMap := headerMap.ParachainHeaderMap.ParachainHeaderMap
		for num := range parachainHeaderMap {
			if latestHeight < num {
				latestHeight = num
			}
		}
		parachainHeader := parachainHeaderMap[latestHeight]
		var decodeHeader gsrpctypes.Header
		err := gsrpccodec.Decode(parachainHeader.BlockHeader, &decodeHeader)
		if err != nil {
			return *latestHeader, latestTimestamp, sdkerrors.Wrapf(err, "decode header error")
		}

		// verify timestamp and get it
		err = beefy.VerifyStateProof(parachainHeader.Timestamp.Proofs,
			decodeHeader.StateRoot[:], parachainHeader.Timestamp.Key,
			parachainHeader.Timestamp.Value)
		if err != nil {
			Logger.Error("LightClient:", "10-Grandpa", "method:", "VerifyStateProof error: ", err)
			return *latestHeader, latestTimestamp, sdkerrors.Wrapf(err, "verify timestamp error")
		}
		//decode
		var decodeTimestamp gsrpctypes.U64
		err = gsrpccodec.Decode(parachainHeader.Timestamp.Value, &decodeTimestamp)
		if err != nil {
			return *latestHeader, latestTimestamp, sdkerrors.Wrapf(err, "decode timestamp error")
		}
		latestTimestamp = time.UnixMilli(int64(decodeTimestamp))

	}

	return *latestHeader, latestTimestamp, nil
}

// ValidateBasic calls the header ValidateBasic function and checks
// with MsgCreateClient
func (h Header) ValidateBasic() error {
	if reflect.DeepEqual(h, Header{}) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "Grandpa header cannot be nil")
	}

	// if h.BeefyMmr.SignedCommitment.Commitment.BlockNumber == 0 {
	// 	return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "latest beefy mmr height cannot be nil")
	// }
	if h.Message == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "Header Message cannot be nil")
	}

	return nil
}
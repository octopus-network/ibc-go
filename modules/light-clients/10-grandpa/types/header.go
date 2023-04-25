package types

import (
	"reflect"
	time "time"

	gsrpctypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"
	gsrpccodec "github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
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
	// latestBeefyHeight := h.BeefyMmr.SignedCommitment.Commitment.BlockNumber
	//get latest header and time
	latestHeader, _, err := getLastestBlockHeader(h)
	if err != nil {
		Logger.Error("LightClient:", "10-Grandpa", "method:", "GetHeight error: ", err)
		return nil
	}
	return clienttypes.NewHeight(0, uint64(latestHeader.Number))
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
	Logger.Debug("latestTime:", latestTime)
	if err != nil {
		Logger.Error("LightClient:", "10-Grandpa", "method:", "getLastestHeader error: ", err)
		return time.Unix(0, 0)
	}
	return latestTime

}

func getLastestBlockHeader(h Header) (gsrpctypes.Header, time.Time, error) {
	var latestBlockHeader gsrpctypes.Header
	var latestTimestamp time.Time
	var latestHeight uint32
	headerMessage := h.GetMessage()

	switch msg := headerMessage.(type) {
	case *Header_SubchainHeaders:
		subchainHeaders := msg.SubchainHeaders.SubchainHeaders

		// convert subchainheaders to subchainheaders map
		subchainHeaderMap := make(map[uint32]SubchainHeader)
		for _, header := range subchainHeaders {
			if latestHeight < header.BlockNumber {
				latestHeight = header.BlockNumber
			}
			subchainHeaderMap[header.BlockNumber]=header
		}

		// find lastest subchain header
		latestSubchainHeader := subchainHeaderMap[latestHeight]

		// var decodeHeader gsrpctypes.Header
		err := gsrpccodec.Decode(latestSubchainHeader.BlockHeader, &latestBlockHeader)
		if err != nil {
			return latestBlockHeader, latestTimestamp, sdkerrors.Wrapf(err, "decode header error")
		}
		Logger.Debug("decodeHeader.Number:", latestBlockHeader.Number)

		// verify timestamp and get it
		err = beefy.VerifyStateProof(latestSubchainHeader.Timestamp.Proofs,
			latestBlockHeader.StateRoot[:], latestSubchainHeader.Timestamp.Key,
			latestSubchainHeader.Timestamp.Value)
		if err != nil {
			Logger.Error("LightClient:", "10-Grandpa", "method:", "VerifyStateProof error: ", err)
			return latestBlockHeader, latestTimestamp, sdkerrors.Wrapf(err, "verify timestamp error")
		}
		//decode
		var decodeTimestamp gsrpctypes.U64
		err = gsrpccodec.Decode(latestSubchainHeader.Timestamp.Value, &decodeTimestamp)
		if err != nil {
			Logger.Error("decode timestamp error:", err)
			return latestBlockHeader, latestTimestamp, sdkerrors.Wrapf(err, "decode timestamp error")
		}
		latestTimestamp = time.UnixMilli(int64(decodeTimestamp))

	case *Header_ParachainHeaders:
		parachainHeaders := msg.ParachainHeaders.ParachainHeaders

		// convert subchainheaders to subchainheaders map
		parachainHeaderMap := make(map[uint32]ParachainHeader)
		for _, header := range parachainHeaders {
			if latestHeight < header.BlockNumber {
				latestHeight = header.BlockNumber
			}
			parachainHeaderMap[header.BlockNumber]=header
		}
		latestParachainHeader := parachainHeaderMap[latestHeight]
		// var decodeHeader gsrpctypes.Header
		err := gsrpccodec.Decode(latestParachainHeader.BlockHeader, &latestBlockHeader)
		if err != nil {
			return latestBlockHeader, latestTimestamp, sdkerrors.Wrapf(err, "decode header error")
		}

		// verify timestamp and get it
		err = beefy.VerifyStateProof(latestParachainHeader.Timestamp.Proofs,
			latestBlockHeader.StateRoot[:], latestParachainHeader.Timestamp.Key,
			latestParachainHeader.Timestamp.Value)
		if err != nil {
			Logger.Error("LightClient:", "10-Grandpa", "method:", "VerifyStateProof error: ", err)
			return latestBlockHeader, latestTimestamp, sdkerrors.Wrapf(err, "verify timestamp error")
		}
		//decode
		var decodeTimestamp gsrpctypes.U64
		err = gsrpccodec.Decode(latestParachainHeader.Timestamp.Value, &decodeTimestamp)
		if err != nil {
			Logger.Error("decode timestamp error:", err)
			return latestBlockHeader, latestTimestamp, sdkerrors.Wrapf(err, "decode timestamp error")
		}
		latestTimestamp = time.UnixMilli(int64(decodeTimestamp))

	}

	return latestBlockHeader, latestTimestamp, nil
}

// ValidateBasic calls the header ValidateBasic function and checks
// with MsgCreateClient
func (h Header) ValidateBasic() error {
	if reflect.DeepEqual(h, Header{}) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "Grandpa header cannot be nil")
	}

	if reflect.DeepEqual(h.BeefyMmr, BeefyMMR{}) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "beefy mmr cannot be nil")
	}

	if h.Message == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "Header Message cannot be nil")
	}

	return nil
}

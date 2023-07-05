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

	//get latest header and time
	latestHeader, _, err := getLastestBlockHeader(h)
	if err != nil {
		return nil
	}
	latestHeight := clienttypes.NewHeight(0, uint64(latestHeader.Number))

	return latestHeight

}

// // GetTime returns the current block timestamp. It returns a zero time if
// // the grandpa header is nil.
// // NOTE: the header.Header is checked to be non nil in ValidateBasic.
func (h Header) GetTime() time.Time {
	//get latest header and time
	_, latestTime, err := getLastestBlockHeader(h)

	if err != nil {
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
			subchainHeaderMap[header.BlockNumber] = header
		}

		// find lastest subchain header
		latestSubchainHeader := subchainHeaderMap[latestHeight]

		// var decodeHeader gsrpctypes.Header
		err := gsrpccodec.Decode(latestSubchainHeader.BlockHeader, &latestBlockHeader)
		if err != nil {
			return latestBlockHeader, latestTimestamp, sdkerrors.Wrapf(err, "decode header error")
		}

		// verify timestamp and get it
		err = beefy.VerifyStateProof(latestSubchainHeader.Timestamp.Proofs,
			latestBlockHeader.StateRoot[:], latestSubchainHeader.Timestamp.Key,
			latestSubchainHeader.Timestamp.Value)
		if err != nil {
			return latestBlockHeader, latestTimestamp, sdkerrors.Wrapf(err, "verify timestamp error")
		}
		//decode
		var decodeTimestamp gsrpctypes.U64
		err = gsrpccodec.Decode(latestSubchainHeader.Timestamp.Value, &decodeTimestamp)
		if err != nil {
			return latestBlockHeader, latestTimestamp, sdkerrors.Wrapf(err, "decode timestamp error")
		}
		latestTimestamp = time.UnixMilli(int64(decodeTimestamp))

	case *Header_ParachainHeaders:
		parachainHeaders := msg.ParachainHeaders.ParachainHeaders
		// convert pb parachainheaders to parachainheaders map
		parachainHeaderMap := make(map[uint32]ParachainHeader)
		for _, header := range parachainHeaders {
			if latestHeight < header.RelayerChainNumber {
				latestHeight = header.RelayerChainNumber
			}
			parachainHeaderMap[header.RelayerChainNumber] = header
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
			return latestBlockHeader, latestTimestamp, sdkerrors.Wrapf(err, "verify timestamp error")
		}
		//decode
		var decodeTimestamp gsrpctypes.U64
		err = gsrpccodec.Decode(latestParachainHeader.Timestamp.Value, &decodeTimestamp)
		if err != nil {
			return latestBlockHeader, latestTimestamp, sdkerrors.Wrapf(err, "decode timestamp error")
		}
		latestTimestamp = time.UnixMilli(int64(decodeTimestamp))

	}

	return latestBlockHeader, latestTimestamp, nil
}

// ValidateBasic calls the header ValidateBasic function and checks
// with MsgCreateClient
func (h Header) ValidateBasic() error {
	if reflect.DeepEqual(h, &Header{}) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "Grandpa header cannot be nil")
	}

	if reflect.DeepEqual(h.BeefyMmr, &BeefyMMR{}) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "beefy mmr cannot be nil")
	}

	if h.Message == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "Header Message cannot be nil")
	}

	return nil
}

package keeper

import (
	"context"
	"fmt"

	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v3/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

// ClientState implements the IBC QueryServer interface
func (q Keeper) ClientState(c context.Context, req *clienttypes.QueryClientStateRequest) (*clienttypes.QueryClientStateResponse, error) {
	fmt.Println("************************* grpc server receive the  query ClientState request ***************************")
	fmt.Println(req)
	return q.ClientKeeper.ClientState(c, req)
}

// ClientStates implements the IBC QueryServer interface
func (q Keeper) ClientStates(c context.Context, req *clienttypes.QueryClientStatesRequest) (*clienttypes.QueryClientStatesResponse, error) {
	fmt.Println("************************* grpc server receive the  query ClientStates request ***************************")
	fmt.Println(req)
	return q.ClientKeeper.ClientStates(c, req)
}

// ConsensusState implements the IBC QueryServer interface
func (q Keeper) ConsensusState(c context.Context, req *clienttypes.QueryConsensusStateRequest) (*clienttypes.QueryConsensusStateResponse, error) {
	fmt.Println("************************* grpc server receive the  query ConsensusState request ***************************")
	fmt.Println(req)
	return q.ClientKeeper.ConsensusState(c, req)
}

// ConsensusStates implements the IBC QueryServer interface
func (q Keeper) ConsensusStates(c context.Context, req *clienttypes.QueryConsensusStatesRequest) (*clienttypes.QueryConsensusStatesResponse, error) {
	fmt.Println("************************* grpc server receive the  query ConsensusStates request ***************************")
	fmt.Println(req)
	return q.ClientKeeper.ConsensusStates(c, req)
}

// ClientStatus implements the IBC QueryServer interface
func (q Keeper) ClientStatus(c context.Context, req *clienttypes.QueryClientStatusRequest) (*clienttypes.QueryClientStatusResponse, error) {
	fmt.Println("************************* grpc server receive the  query ClientStatus request ***************************")
	fmt.Println(req)
	return q.ClientKeeper.ClientStatus(c, req)
}

// ClientParams implements the IBC QueryServer interface
func (q Keeper) ClientParams(c context.Context, req *clienttypes.QueryClientParamsRequest) (*clienttypes.QueryClientParamsResponse, error) {
	fmt.Println("************************* grpc server receive the  query ClientParams request ***************************")
	fmt.Println(req)
	return q.ClientKeeper.ClientParams(c, req)
}

// UpgradedClientState implements the IBC QueryServer interface
func (q Keeper) UpgradedClientState(c context.Context, req *clienttypes.QueryUpgradedClientStateRequest) (*clienttypes.QueryUpgradedClientStateResponse, error) {
	fmt.Println("************************* grpc server receive the  query UpgradedClientState request ***************************")
	fmt.Println(req)
	return q.ClientKeeper.UpgradedClientState(c, req)
}

// Connection implements the IBC QueryServer interface
func (q Keeper) Connection(c context.Context, req *connectiontypes.QueryConnectionRequest) (*connectiontypes.QueryConnectionResponse, error) {
	fmt.Println("************************* grpc server receive the  query Connection request ***************************")
	fmt.Println(req)
	return q.ConnectionKeeper.Connection(c, req)
}

// Connections implements the IBC QueryServer interface
func (q Keeper) Connections(c context.Context, req *connectiontypes.QueryConnectionsRequest) (*connectiontypes.QueryConnectionsResponse, error) {
	fmt.Println("************************* grpc server receive the  query Connections request ***************************")
	fmt.Println(req)
	return q.ConnectionKeeper.Connections(c, req)
}

// ClientConnections implements the IBC QueryServer interface
func (q Keeper) ClientConnections(c context.Context, req *connectiontypes.QueryClientConnectionsRequest) (*connectiontypes.QueryClientConnectionsResponse, error) {
	fmt.Println("************************* grpc server receive the  query ClientConnections request ***************************")
	fmt.Println(req)
	return q.ConnectionKeeper.ClientConnections(c, req)
}

// ConnectionClientState implements the IBC QueryServer interface
func (q Keeper) ConnectionClientState(c context.Context, req *connectiontypes.QueryConnectionClientStateRequest) (*connectiontypes.QueryConnectionClientStateResponse, error) {
	fmt.Println("************************* grpc server receive the  query ConnectionClientState request ***************************")
	fmt.Println(req)
	return q.ConnectionKeeper.ConnectionClientState(c, req)
}

// ConnectionConsensusState implements the IBC QueryServer interface
func (q Keeper) ConnectionConsensusState(c context.Context, req *connectiontypes.QueryConnectionConsensusStateRequest) (*connectiontypes.QueryConnectionConsensusStateResponse, error) {
	fmt.Println("************************* grpc server receive the  query ConnectionConsensusState request ***************************")
	fmt.Println(req)
	return q.ConnectionKeeper.ConnectionConsensusState(c, req)
}

// Channel implements the IBC QueryServer interface
func (q Keeper) Channel(c context.Context, req *channeltypes.QueryChannelRequest) (*channeltypes.QueryChannelResponse, error) {
	fmt.Println("************************* grpc server receive the  query Channel request ***************************")
	fmt.Println(req)
	return q.ChannelKeeper.Channel(c, req)
}

// Channels implements the IBC QueryServer interface
func (q Keeper) Channels(c context.Context, req *channeltypes.QueryChannelsRequest) (*channeltypes.QueryChannelsResponse, error) {
	fmt.Println("************************* grpc server receive the  query Channels request ***************************")
	fmt.Println(req)
	return q.ChannelKeeper.Channels(c, req)
}

// ConnectionChannels implements the IBC QueryServer interface
func (q Keeper) ConnectionChannels(c context.Context, req *channeltypes.QueryConnectionChannelsRequest) (*channeltypes.QueryConnectionChannelsResponse, error) {
	fmt.Println("************************* grpc server receive the  query ConnectionChannels request ***************************")
	fmt.Println(req)
	return q.ChannelKeeper.ConnectionChannels(c, req)
}

// ChannelClientState implements the IBC QueryServer interface
func (q Keeper) ChannelClientState(c context.Context, req *channeltypes.QueryChannelClientStateRequest) (*channeltypes.QueryChannelClientStateResponse, error) {
	fmt.Println("************************* grpc server receive the  query ChannelClientState request ***************************")
	fmt.Println(req)
	return q.ChannelKeeper.ChannelClientState(c, req)
}

// ChannelConsensusState implements the IBC QueryServer interface
func (q Keeper) ChannelConsensusState(c context.Context, req *channeltypes.QueryChannelConsensusStateRequest) (*channeltypes.QueryChannelConsensusStateResponse, error) {
	fmt.Println("************************* grpc server receive the  query ChannelConsensusState request ***************************")
	fmt.Println(req)
	return q.ChannelKeeper.ChannelConsensusState(c, req)
}

// PacketCommitment implements the IBC QueryServer interface
func (q Keeper) PacketCommitment(c context.Context, req *channeltypes.QueryPacketCommitmentRequest) (*channeltypes.QueryPacketCommitmentResponse, error) {
	fmt.Println("************************* grpc server receive the  query PacketCommitment request ***************************")
	fmt.Println(req)
	return q.ChannelKeeper.PacketCommitment(c, req)
}

// PacketCommitments implements the IBC QueryServer interface
func (q Keeper) PacketCommitments(c context.Context, req *channeltypes.QueryPacketCommitmentsRequest) (*channeltypes.QueryPacketCommitmentsResponse, error) {
	fmt.Println("************************* grpc server receive the  query PacketCommitments request ***************************")
	fmt.Println(req)
	return q.ChannelKeeper.PacketCommitments(c, req)
}

// PacketReceipt implements the IBC QueryServer interface
func (q Keeper) PacketReceipt(c context.Context, req *channeltypes.QueryPacketReceiptRequest) (*channeltypes.QueryPacketReceiptResponse, error) {
	fmt.Println("************************* grpc server receive the  query PacketReceipt request ***************************")
	fmt.Println(req)
	return q.ChannelKeeper.PacketReceipt(c, req)
}

// PacketAcknowledgement implements the IBC QueryServer interface
func (q Keeper) PacketAcknowledgement(c context.Context, req *channeltypes.QueryPacketAcknowledgementRequest) (*channeltypes.QueryPacketAcknowledgementResponse, error) {
	fmt.Println("************************* grpc server receive the  query PacketAcknowledgement request ***************************")
	fmt.Println(req)
	return q.ChannelKeeper.PacketAcknowledgement(c, req)
}

// PacketAcknowledgements implements the IBC QueryServer interface
func (q Keeper) PacketAcknowledgements(c context.Context, req *channeltypes.QueryPacketAcknowledgementsRequest) (*channeltypes.QueryPacketAcknowledgementsResponse, error) {
	fmt.Println("************************* grpc server receive the  query PacketAcknowledgements request ***************************")
	fmt.Println(req)
	return q.ChannelKeeper.PacketAcknowledgements(c, req)
}

// UnreceivedPackets implements the IBC QueryServer interface
func (q Keeper) UnreceivedPackets(c context.Context, req *channeltypes.QueryUnreceivedPacketsRequest) (*channeltypes.QueryUnreceivedPacketsResponse, error) {
	fmt.Println("************************* grpc server receive the  query UnreceivedPackets request ***************************")
	fmt.Println(req)
	return q.ChannelKeeper.UnreceivedPackets(c, req)
}

// UnreceivedAcks implements the IBC QueryServer interface
func (q Keeper) UnreceivedAcks(c context.Context, req *channeltypes.QueryUnreceivedAcksRequest) (*channeltypes.QueryUnreceivedAcksResponse, error) {
	fmt.Println("************************* grpc server receive the  query UnreceivedAcks request ***************************")
	fmt.Println(req)
	return q.ChannelKeeper.UnreceivedAcks(c, req)
}

// NextSequenceReceive implements the IBC QueryServer interface
func (q Keeper) NextSequenceReceive(c context.Context, req *channeltypes.QueryNextSequenceReceiveRequest) (*channeltypes.QueryNextSequenceReceiveResponse, error) {
	fmt.Println("************************* grpc server receive the  query NextSequenceReceive request ***************************")
	fmt.Println(req)
	return q.ChannelKeeper.NextSequenceReceive(c, req)
}

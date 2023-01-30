package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
)

var _ types.MsgServer = Keeper{}

// Transfer defines a rpc handler method for MsgTransfer.
func (k Keeper) Transfer(goCtx context.Context, msg *types.MsgTransfer) (*types.MsgTransferResponse, error) {
	fmt.Println("[msg_server]************************* grpc server receive the  Transfer begin ***************************")

	ctx := sdk.UnwrapSDKContext(goCtx)

	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, err
	}

	sequence, err := k.sendTransfer(
		ctx, msg.SourcePort, msg.SourceChannel, msg.Token, sender, msg.Receiver, msg.TimeoutHeight, msg.TimeoutTimestamp,
		msg.Memo)
	if err != nil {
		return nil, err
	}

	k.Logger(ctx).Info("IBC fungible token transfer", "token", msg.Token.Denom, "amount", msg.Token.Amount.String(), "sender", msg.Sender, "receiver", msg.Receiver)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeTransfer,
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Sender),
			sdk.NewAttribute(types.AttributeKeyReceiver, msg.Receiver),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	})

	fmt.Println("[msg_server]************************* grpc server receive the  CreateClient end ***************************")
	return &types.MsgTransferResponse{Sequence: sequence}, nil

}

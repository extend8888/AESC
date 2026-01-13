package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Message types for the execution module
const (
	TypeMsgBatchIngest = "batch_ingest"
)

var (
	_ sdk.Msg = &MsgBatchIngest{}
)

// Route implements sdk.Msg
func (msg MsgBatchIngest) Route() string { return RouterKey }

// Type implements sdk.Msg
func (msg MsgBatchIngest) Type() string { return TypeMsgBatchIngest }

// GetSignBytes implements sdk.Msg
func (msg MsgBatchIngest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&msg)
	return sdk.MustSortJSON(bz)
}

// GetSigners implements sdk.Msg
func (msg MsgBatchIngest) GetSigners() []sdk.AccAddress {
	sender, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{sender}
}

// ValidateBasic implements sdk.Msg
func (msg MsgBatchIngest) ValidateBasic() error {
	// Validate sender address
	_, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sender address: %s", err)
	}

	// Validate batch_id
	if len(msg.BatchId) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "batch_id cannot be empty")
	}

	// Validate source_data
	if len(msg.SourceData) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "source_data cannot be empty")
	}

	return nil
}

package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/execution module sentinel errors
var (
	// ErrOrderAlreadyExists is returned when attempting to create an order that already exists
	ErrOrderAlreadyExists = sdkerrors.Register(ModuleName, 2, "order already exists")

	// ErrOrderNotFound is returned when the requested order does not exist
	ErrOrderNotFound = sdkerrors.Register(ModuleName, 3, "order not found")

	// ErrInvalidPair is returned when the trading pair is invalid or empty
	ErrInvalidPair = sdkerrors.Register(ModuleName, 4, "invalid trading pair")

	// ErrInvalidOrderId is returned when the order ID is invalid or empty
	ErrInvalidOrderId = sdkerrors.Register(ModuleName, 5, "invalid order ID")

	// ErrInvalidOwner is returned when the owner address is invalid or empty
	ErrInvalidOwner = sdkerrors.Register(ModuleName, 6, "invalid owner address")

	// ErrInvalidSide is returned when the order side is not "buy" or "sell"
	ErrInvalidSide = sdkerrors.Register(ModuleName, 7, "invalid order side, must be 'buy' or 'sell'")

	// ErrInvalidPrice is returned when the order price is invalid (zero or negative)
	ErrInvalidPrice = sdkerrors.Register(ModuleName, 8, "invalid order price, must be positive")

	// ErrInvalidQuantity is returned when the order quantity is invalid (zero or negative)
	ErrInvalidQuantity = sdkerrors.Register(ModuleName, 9, "invalid order quantity, must be positive")

	// ErrInvalidOrderType is returned when the order type is invalid or empty
	ErrInvalidOrderType = sdkerrors.Register(ModuleName, 10, "invalid order type")

	// ErrPairNotWhitelisted is returned when the trading pair is not in the whitelist
	ErrPairNotWhitelisted = sdkerrors.Register(ModuleName, 11, "trading pair not whitelisted")
)

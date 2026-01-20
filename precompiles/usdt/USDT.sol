// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

address constant USDT_PRECOMPILE_ADDRESS = 0x0000000000000000000000000000000000001010;

IUSDT constant USDT_CONTRACT = IUSDT(
    USDT_PRECOMPILE_ADDRESS
);

interface IUSDT {
    // Events
    event Transfer(address indexed from, address indexed to, uint256 value);
    event Approval(address indexed owner, address indexed spender, uint256 value);
    event AuthorizationUsed(address indexed authorizer, bytes32 indexed nonce);
    event AuthorizationCanceled(address indexed authorizer, bytes32 indexed nonce);

    // ERC-20 Queries
    function name() external view returns (string memory);

    function symbol() external view returns (string memory);

    function decimals() external view returns (uint8);

    function totalSupply() external view returns (uint256);

    function balanceOf(
        address account
    ) external view returns (uint256);

    function allowance(
        address owner,
        address spender
    ) external view returns (uint256);

    // ERC-20 Transactions
    function transfer(
        address to,
        uint256 amount
    ) external returns (bool);

    function approve(
        address spender,
        uint256 amount
    ) external returns (bool);

    function transferFrom(
        address from,
        address to,
        uint256 amount
    ) external returns (bool);

    // EIP-712
    function DOMAIN_SEPARATOR() external view returns (bytes32);

    // EIP-3009 Queries
    function authorizationState(
        address authorizer,
        bytes32 nonce
    ) external view returns (bool);

    // EIP-3009 Transactions
    function transferWithAuthorization(
        address from,
        address to,
        uint256 value,
        uint256 validAfter,
        uint256 validBefore,
        bytes32 nonce,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external;

    function cancelAuthorization(
        address authorizer,
        bytes32 nonce,
        uint8 v,
        bytes32 r,
        bytes32 s
    ) external;
}


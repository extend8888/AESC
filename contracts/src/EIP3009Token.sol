// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/utils/cryptography/EIP712.sol";

/**
 * @title EIP3009Token
 * @dev ERC20 token with EIP-3009 transferWithAuthorization support
 * This contract is used for testing x402-relayer with custom ERC20 tokens
 */
contract EIP3009Token is ERC20, EIP712 {
    using ECDSA for bytes32;

    // EIP-3009 type hash
    bytes32 public constant TRANSFER_WITH_AUTHORIZATION_TYPEHASH = keccak256(
        "TransferWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)"
    );

    // Track used authorization nonces
    mapping(address => mapping(bytes32 => bool)) private _authorizationStates;

    // Events
    event AuthorizationUsed(address indexed authorizer, bytes32 indexed nonce);

    constructor(
        string memory name_,
        string memory symbol_,
        uint256 initialSupply
    ) ERC20(name_, symbol_) EIP712(name_, "1") {
        _mint(msg.sender, initialSupply);
    }

    /**
     * @dev Returns the domain separator used in EIP-712 signatures
     */
    function DOMAIN_SEPARATOR() external view returns (bytes32) {
        return _domainSeparatorV4();
    }

    /**
     * @dev Returns the state of an authorization nonce
     * @param authorizer The authorizer address
     * @param nonce The nonce to check
     * @return True if the nonce has been used
     */
    function authorizationState(address authorizer, bytes32 nonce) external view returns (bool) {
        return _authorizationStates[authorizer][nonce];
    }

    /**
     * @dev Execute a transfer with a signed authorization (EIP-3009)
     * @param from The token holder address
     * @param to The recipient address
     * @param value The amount to transfer
     * @param validAfter The time after which the authorization is valid
     * @param validBefore The time before which the authorization is valid
     * @param nonce A unique nonce for this authorization
     * @param v ECDSA signature v
     * @param r ECDSA signature r
     * @param s ECDSA signature s
     */
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
    ) external {
        require(block.timestamp > validAfter, "EIP3009: authorization not yet valid");
        require(block.timestamp < validBefore, "EIP3009: authorization expired");
        require(!_authorizationStates[from][nonce], "EIP3009: authorization already used");

        bytes32 structHash = keccak256(
            abi.encode(
                TRANSFER_WITH_AUTHORIZATION_TYPEHASH,
                from,
                to,
                value,
                validAfter,
                validBefore,
                nonce
            )
        );

        bytes32 digest = _hashTypedDataV4(structHash);
        address signer = ECDSA.recover(digest, v, r, s);
        require(signer == from, "EIP3009: invalid signature");

        _authorizationStates[from][nonce] = true;
        emit AuthorizationUsed(from, nonce);

        _transfer(from, to, value);
    }

    /**
     * @dev Mint tokens to an address (for testing purposes)
     * @param to The recipient address
     * @param amount The amount to mint
     */
    function mint(address to, uint256 amount) external {
        _mint(to, amount);
    }

    /**
     * @dev Returns the token decimals (18 for standard ERC20 tokens)
     */
    function decimals() public pure override returns (uint8) {
        return 18;
    }
}


package app_test

import (
	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	testkeeper "github.com/sei-protocol/sei-chain/testutil/keeper"
)

func signTx(txBuilder client.TxBuilder, privKey cryptotypes.PrivKey, acc authtypes.AccountI) sdk.Tx {
	var sigsV2 []signing.SignatureV2
	sigV2 := signing.SignatureV2{
		PubKey: privKey.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode:  testkeeper.EVMTestApp.GetTxConfig().SignModeHandler().DefaultMode(),
			Signature: nil,
		},
		Sequence: acc.GetSequence(),
	}
	sigsV2 = append(sigsV2, sigV2)
	_ = txBuilder.SetSignatures(sigsV2...)
	sigsV2 = []signing.SignatureV2{}
	signerData := xauthsigning.SignerData{
		ChainID:       "sei-test",
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      acc.GetSequence(),
	}
	sigV2, _ = clienttx.SignWithPrivKey(
		testkeeper.EVMTestApp.GetTxConfig().SignModeHandler().DefaultMode(),
		signerData,
		txBuilder,
		privKey,
		testkeeper.EVMTestApp.GetTxConfig(),
		acc.GetSequence(),
	)
	sigsV2 = append(sigsV2, sigV2)
	_ = txBuilder.SetSignatures(sigsV2...)
	return txBuilder.GetTx()
}

func signTxMultiple(txBuilder client.TxBuilder, privKeys []cryptotypes.PrivKey, accs []authtypes.AccountI) sdk.Tx {
	var sigsV2 []signing.SignatureV2
	for i, privKey := range privKeys {
		sigsV2 = append(sigsV2, signing.SignatureV2{
			PubKey: privKey.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  testkeeper.EVMTestApp.GetTxConfig().SignModeHandler().DefaultMode(),
				Signature: nil,
			},
			Sequence: accs[i].GetSequence(),
		})
	}
	_ = txBuilder.SetSignatures(sigsV2...)
	sigsV2 = []signing.SignatureV2{}
	for i, privKey := range privKeys {
		signerData := xauthsigning.SignerData{
			ChainID:       "sei-test",
			AccountNumber: accs[i].GetAccountNumber(),
			Sequence:      accs[i].GetSequence(),
		}
		sigV2, _ := clienttx.SignWithPrivKey(
			testkeeper.EVMTestApp.GetTxConfig().SignModeHandler().DefaultMode(),
			signerData,
			txBuilder,
			privKey,
			testkeeper.EVMTestApp.GetTxConfig(),
			accs[i].GetSequence(),
		)
		sigsV2 = append(sigsV2, sigV2)
	}
	_ = txBuilder.SetSignatures(sigsV2...)
	return txBuilder.GetTx()
}


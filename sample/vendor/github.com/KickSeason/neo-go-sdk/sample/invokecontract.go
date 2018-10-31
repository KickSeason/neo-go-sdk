package main

import (
	"encoding/hex"
	"log"

	"github.com/KickSeason/neo-go-sdk/neocliapi"
	"github.com/KickSeason/neo-go-sdk/neotransaction"
)

var (
	contractHashString = `9e04d7fddc770cb422b2781c6d84771d4d34cd7a`
	addrReceiverString = "AUpVUWUjP15zXk6MAB4uhYRYJpXnyphxtp"
	neorpcurl          = "http://47.98.227.225:50332"
)

// InvokeContract 调用一个智能合约的 transfer 接口
func InvokeContract() {

	contractHash, _ := hex.DecodeString(contractHashString)

	// The key used to sign the transaction if needed
	key, _ := neotransaction.DecodeFromWif("L3MgmkFsvU5WUL8bJhUeDYFzbvHkPEJTybNxwecJU6yX8oks42V4")
	addr := key.CreateBasicAddress()
	addrReceiver, _ := neotransaction.ParseAddress(addrReceiverString)
	// 创建一个 Invocation 交易
	tx := neotransaction.CreateInvocationTransaction()

	extra := tx.ExtraData.(*neotransaction.InvocationExtraData)

	args := []interface{}{
		addr.ScripHash,
		addrReceiver.ScripHash,
		1,
	}
	bytes, err := neotransaction.BuildCallMethodScript(contractHash, "transfer", args, true)
	if err != nil {
		log.Println(err)
	}
	extra.Script = bytes

	// If the transaction need additional Witness then put the ScriptHash in attributes
	tx.AppendAttribute(neotransaction.UsageScript, addr.ScripHash)

	// Perhaps the transaction need Witness
	tx.AppendBasicSignWitness(key)

	log.Printf(`Generate invocation transaction[%s]`, tx.TXID())
	rawtx := tx.RawTransactionString()
	log.Println("transaction content: ", rawtx)

	result := neocliapi.SendRawTransaction(neorpcurl, rawtx)
	log.Printf(`Send transaction to neo-cli node result[%v]`, result)
}

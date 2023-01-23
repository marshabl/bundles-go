package main

import (
	"bundles"
	"bundles/internal"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/marshabl/blocknative-go-sdk/bnsdk"
)

const SYSTEM = "ethereum"
const NETWORK = 1
const MONITORADDRESS = "0xEf1c6E67703c7BD7107eed8303Fbe6EC2554BF6B" //uniswap autorouter
const RPCURL = "https://api.blocknative.com/v1/auction"             //"https://relay.flashbots.net"

type Payload struct {
	Event struct {
		Transaction internal.MpexTransaction
	}
}

func txnHandler(e []byte) {
	var event = new(Payload)
	err := json.Unmarshal(e, event)
	if err != nil {
		log.Printf("failed to parse testcase: %v", err)
	}
	if event.Event.Transaction.Nonce != nil {
		tx, err := internal.BuildTxFromMpex(&event.Event.Transaction)
		if err != nil {
			log.Printf("failed to build tx: %v", err)
		}

		blockNumber := event.Event.Transaction.PendingBlockNumber + 1
		blockNumberHex := internal.IntToHex(blockNumber)
		rawTx := "0x" + internal.TxToRlp(tx)[6:] //ignore first 6 characters - https://ethereum.org/en/developers/docs/data-structures-and-encoding/rlp/
		fmt.Println(rawTx)
		fmt.Println(event.Event.Transaction.Hash)
		var privateKey, _ = crypto.GenerateKey()
		rpc := bundles.NewRPC(RPCURL)
		opts := bundles.CallBundleParam{
			Txs:              []string{rawTx},
			BlockNumber:      blockNumberHex,
			StateBlockNumber: "latest",
		}

		result, err := rpc.CallBundle(privateKey, opts)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Print result
		fmt.Printf("%+v\n", result)
	}

}

func main() {
	APIKEY, ok := os.LookupEnv("APIKEY")
	if !ok {
		log.Printf("No environment variable APIRKEY.")
		return
	}

	var filters []map[string]string
	filter := map[string]string{
		"status": "pending",
	}
	filters = append(filters, filter)
	config := bnsdk.NewConfig("global", filters)

	sub := bnsdk.NewSubscription(common.HexToAddress(MONITORADDRESS), config)
	client, err := bnsdk.Stream(APIKEY, SYSTEM, NETWORK)
	if err != nil {
		log.Printf("error: %v", err.Error())
		return
	}
	client.Subscribe(sub, txnHandler)
}

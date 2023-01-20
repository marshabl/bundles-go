package main

import (
	"bundles"
	"bundles/internal"
	"bytes"
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

		var buff bytes.Buffer
		tx.EncodeRLP(&buff)

		rawTx := fmt.Sprintf("0x%x", buff.Bytes())
		fmt.Println(rawTx)
		var privateKey, _ = crypto.GenerateKey()
		rpc := bundles.NewRPC("https://api.blocknative.com/v1/auction")
		opts := bundles.CallBundleParam{
			Txs:              []string{rawTx},
			BlockNumber:      "0xfaf049",
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

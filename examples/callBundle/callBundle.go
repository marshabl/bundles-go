package main

import (
	"bundles"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	// PRIVATEKEY, ok := os.LookupEnv("PRIVATEKEY")
	// if !ok {
	// 	fmt.Printf("No environment variable PRIVATERKEY.")
	// 	return
	// }

	// var privateKey, _ = crypto.HexToECDSA(PRIVATEKEY)
	var privateKey, _ = crypto.GenerateKey()

	rpc := bundles.NewRPC("https://api.blocknative.com/v1/auction")
	opts := bundles.CallBundleParam{
		Txs:              []string{"<rawTxHash>"},
		BlockNumber:      "<blockNumber>",
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

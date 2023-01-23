package bundles

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// most types pulled from https://github.com/metachris/flashbotsrpc/blob/master/types.go
type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RelayErrorResponse struct {
	Error string `json:"error"`
}

func (err RpcError) Error() string {
	return fmt.Sprintf("Error %d (%s)", err.Code, err.Message)
}

type CallBundleParam struct {
	Txs              []string `json:"txs"`                 // Array[String], A list of signed transactions to execute in an atomic bundle
	BlockNumber      string   `json:"blockNumber"`         // String, a hex encoded block number for which this bundle is valid on
	StateBlockNumber string   `json:"stateBlockNumber"`    // String, either a hex encoded number or a block tag for which state to base this simulation on. Can use "latest"
	Timestamp        int64    `json:"timestamp,omitempty"` // Number, the timestamp to use for this bundle simulation, in seconds since the unix epoch
	Timeout          int64    `json:"timeout,omitempty"`
	GasLimit         uint64   `json:"gasLimit,omitempty"`
	Difficulty       uint64   `json:"difficulty,omitempty"`
	BaseFee          uint64   `json:"baseFee,omitempty"`
}

type CallBundleResult struct {
	CoinbaseDiff      string `json:"coinbaseDiff"`      // "2717471092204423",
	EthSentToCoinbase string `json:"ethSentToCoinbase"` // "0",
	FromAddress       string `json:"fromAddress"`       // "0x37ff310ab11d1928BB70F37bC5E3cf62Df09a01c",
	GasFees           string `json:"gasFees"`           // "2717471092204423",
	GasPrice          string `json:"gasPrice"`          // "43000001459",
	GasUsed           int64  `json:"gasUsed"`           // 63197,
	ToAddress         string `json:"toAddress"`         // "0xdAC17F958D2ee523a2206206994597C13D831ec7",
	TxHash            string `json:"txHash"`            // "0xe2df005210bdc204a34ff03211606e5d8036740c686e9fe4e266ae91cf4d12df",
	Value             string `json:"value"`             // "0x"
	Error             string `json:"error"`
	Revert            string `json:"revert"`
}

type CallBundleResponse struct {
	BundleGasPrice    string             `json:"bundleGasPrice"`    // "43000001459",
	BundleHash        string             `json:"bundleHash"`        // "0x2ca9c4d2ba00d8144d8e396a4989374443cb20fb490d800f4f883ad4e1b32158",
	CoinbaseDiff      string             `json:"coinbaseDiff"`      // "2717471092204423",
	EthSentToCoinbase string             `json:"ethSentToCoinbase"` // "0",
	GasFees           string             `json:"gasFees"`           // "2717471092204423",
	Results           []CallBundleResult `json:"results"`           // [],
	StateBlockNumber  int64              `json:"stateBlockNumber"`  // 12960319,
	TotalGasUsed      int64              `json:"totalGasUsed"`      // 63197
}

type SendBundleRequest struct {
	Txs          []string  `json:"txs"`                         // Array[String], A list of signed transactions to execute in an atomic bundle
	BlockNumber  string    `json:"blockNumber"`                 // String, a hex encoded block number for which this bundle is valid on
	MinTimestamp *uint64   `json:"minTimestamp,omitempty"`      // (Optional) Number, the minimum timestamp for which this bundle is valid, in seconds since the unix epoch
	MaxTimestamp *uint64   `json:"maxTimestamp,omitempty"`      // (Optional) Number, the maximum timestamp for which this bundle is valid, in seconds since the unix epoch
	RevertingTxs *[]string `json:"revertingTxHashes,omitempty"` // (Optional) Array[String], A list of tx hashes that are allowed to revert
}

type rpcResponse struct {
	ID      int             `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *RpcError       `json:"error"`
}

type rpcRequest struct {
	ID      int           `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// FlashbotsRPC - Ethereum rpc client
type RPC struct {
	url         string
	client      httpClient
	Headers     map[string]string
	Signature   string
	RequestBody []byte
	Request     *http.Request
	Timeout     time.Duration
}

// New create new rpc client with given url
func NewRPC(url string) *RPC {
	rpc := &RPC{
		url:     url,
		Headers: make(map[string]string),
		Timeout: 30 * time.Second,
	}
	rpc.client = &http.Client{
		Timeout: rpc.Timeout,
	}
	return rpc
}

func (rpc *RPC) newRPCRequest(method string, params ...interface{}) rpcRequest {
	return rpcRequest{
		ID:      1,
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

func (rpc *RPC) buildSignature(method string, privKey *ecdsa.PrivateKey, params ...interface{}) error {
	request := rpc.newRPCRequest(method, params...)

	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	hashedBody := crypto.Keccak256Hash([]byte(body)).Hex()
	sig, err := crypto.Sign(accounts.TextHash([]byte(hashedBody)), privKey)
	if err != nil {
		return err
	}

	rpc.Signature = crypto.PubkeyToAddress(privKey.PublicKey).Hex() + ":" + hexutil.Encode(sig)
	rpc.RequestBody = body
	return nil
}

func (rpc *RPC) buildRPCRequest() error {
	req, err := http.NewRequest("POST", rpc.url, bytes.NewBuffer(rpc.RequestBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Auction-Signature", rpc.Signature)   //sould be X-Flashbots-Signature for other builders like Flashbots
	req.Header.Add("X-Flashbots-Signature", rpc.Signature) //sould be X-Flashbots-Signature for other builders like Flashbots
	for k, v := range rpc.Headers {
		req.Header.Add(k, v)
	}
	rpc.Request = req
	return nil
}

// func GetMyRawTx()                                                 {}
// func (rpc *RPC) SendBundle(privKey *ecdsa.PrivateKey, params ...interface{}) {}
func (rpc *RPC) CallBundle(privKey *ecdsa.PrivateKey, params ...interface{}) (*CallBundleResponse, error) {
	response, err := rpc.makeRpcCall("eth_callBundle", privKey, params...)
	if err != nil {
		return nil, err
	}

	var cbResponse = new(CallBundleResponse)
	err = json.Unmarshal(response.Result, &cbResponse)
	if err != nil {
		return nil, err
	}
	return cbResponse, nil
}

// CallWithFlashbotsSignature is like Call but also signs the request
func (rpc *RPC) makeRpcCall(method string, privKey *ecdsa.PrivateKey, params ...interface{}) (*rpcResponse, error) {
	if err := rpc.buildSignature(method, privKey, params...); err != nil {
		return nil, err
	}

	if err := rpc.buildRPCRequest(); err != nil {
		return nil, err
	}

	// fmt.Println(rpc)

	response, err := rpc.client.Do(rpc.Request)
	// fmt.Println(response)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	// errorResp := new(RelayErrorResponse)
	// if err := json.Unmarshal(data, errorResp); err == nil && errorResp.Error != "" {
	// 	// relay returned an error
	// 	return nil, fmt.Errorf("%w: %s", ErrRelayErrorResponse, errorResp.Error)
	// }

	resp := new(rpcResponse)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	return resp, nil
}

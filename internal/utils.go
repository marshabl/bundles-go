package internal

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type MpexTransaction struct {
	Nonce                *big.Int `json:"nonce"`
	MaxPriorityFeePerGas string   `json:"maxPriorityFeePerGas,omitempty"`
	MaxFeePerGas         string   `json:"maxFeePerGas,omitempty"`
	Gas                  uint64   `json:"gas"`
	GasPrice             string   `json:"gasPrice,omitempty"`
	To                   string   `json:"to"`
	Value                string   `json:"value"`
	Input                string   `json:"input"`
	V                    string   `json:"v"`
	R                    string   `json:"r"`
	S                    string   `json:"s"`
	Type                 int      `json:"type"`
	Hash                 string   `json:"hash"`
}

func convertStringToBigInt(s string, base int) (*big.Int, error) {
	ret := new(big.Int)
	ret, ok := ret.SetString(s, base)
	if !ok {
		return nil, fmt.Errorf("failed to set string %s to base %v", s, base)
	}
	return ret, nil
}

func BuildTxFromMpex(m *MpexTransaction) (*types.Transaction, error) {
	chainId := big.NewInt(1)

	nonce := m.Nonce.Uint64()

	gas := m.Gas

	to := common.HexToAddress(m.To)

	value, err := convertStringToBigInt(m.Value, 10)
	if err != nil {
		return nil, err
	}

	data, err := hex.DecodeString(m.Input[2:])
	if err != nil {
		return nil, err
	}

	v, err := convertStringToBigInt(m.V[2:], 16)
	if err != nil {
		return nil, err
	}

	r, err := convertStringToBigInt(m.R[2:], 16)
	if err != nil {
		return nil, err
	}

	s, err := convertStringToBigInt(m.S[2:], 16)
	if err != nil {
		return nil, err
	}

	if m.Type == 2 {
		gasTipCap, err := convertStringToBigInt(m.MaxPriorityFeePerGas, 10)
		if err != nil {
			return nil, err
		}

		gasFeeCap, err := convertStringToBigInt(m.MaxFeePerGas, 10)
		if err != nil {
			return nil, err
		}

		dynamicTx := types.DynamicFeeTx{
			ChainID:   chainId,
			Nonce:     nonce,
			GasTipCap: gasTipCap,
			GasFeeCap: gasFeeCap,
			Gas:       gas,
			To:        &to,
			Value:     value,
			Data:      data,
			V:         v,
			R:         r,
			S:         s,
		}
		return types.NewTx(&dynamicTx), nil
	}

	if m.Type == 0 {
		gasPrice, err := convertStringToBigInt(m.GasPrice, 10)
		if err != nil {
			return nil, err
		}

		legacyTx := types.LegacyTx{
			Nonce:    nonce,
			GasPrice: gasPrice,
			Gas:      gas,
			To:       &to,
			Value:    value,
			Data:     data,
			V:        v,
			R:        r,
			S:        s,
		}
		return types.NewTx(&legacyTx), nil
	}

	return nil, fmt.Errorf("unable to convert MPEX transaction to GETH transaction type")

}

package encodetest

import (
	"bundles/internal"
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
)

func TestEncodingMpex(t *testing.T) {
	testEncodingMpex(t)
}

func testEncodingMpex(t *testing.T) {
	t.Parallel()
	client, err := ethclient.Dial("wss://necessary-newest-waterfall.quiknode.pro/048d029a37818e6a8dfb4dc4eeeebc8db889913e/")
	if err != nil {
		log.Fatal(err)
	}
	files, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatalf("failed to retrieve tracer test suite: %v", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		file := file

		var mpexTx = new(internal.MpexTransaction)

		if blob, err := os.ReadFile(filepath.Join("testdata", file.Name())); err != nil {
			t.Fatalf("failed to read testcase: %v", err)
		} else if err := json.Unmarshal(blob, mpexTx); err != nil {
			t.Fatalf("failed to parse testcase: %v", err)
		}

		tx, err := internal.BuildTxFromMpex(mpexTx)
		if err != nil {
			t.Fatalf("failed to build tx: %v", err)
		}

		hash := common.HexToHash(mpexTx.Hash)
		w3Tx, _, err := client.TransactionByHash(context.Background(), hash)
		if err != nil {
			t.Fatalf("w3 client failed to get transaction from hash: %v", err)
		}

		var w3Buff bytes.Buffer
		w3Tx.EncodeRLP(&w3Buff)

		var buff bytes.Buffer
		tx.EncodeRLP(&buff)

		assert.Equal(t, w3Buff.Bytes(), buff.Bytes())
		assert.Equal(t, w3Tx.Hash(), tx.Hash())
	}
}

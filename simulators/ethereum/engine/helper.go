package main

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/hive/hivesim"
)

// default timeout for RPC calls
var rpcTimeout = 10 * time.Second

// Retrieves contract storage as BigInt
func getBigIntAtStorage(eth *ethclient.Client, ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) (*big.Int, error) {
	stor, err := eth.StorageAt(ctx, account, key, blockNumber)
	if err != nil {
		return nil, err
	}
	bigint := big.NewInt(0)
	bigint.SetBytes(stor)
	return bigint, nil
}

// From ethereum/rpc:

// loggingRoundTrip writes requests and responses to the test log.
type loggingRoundTrip struct {
	t     *hivesim.T
	hc    *hivesim.Client
	inner http.RoundTripper
}

func (rt *loggingRoundTrip) RoundTrip(req *http.Request) (*http.Response, error) {
	// Read and log the request body.
	reqBytes, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return nil, err
	}
	rt.t.Logf(">> (%s) %s", rt.hc.Container, bytes.TrimSpace(reqBytes))
	reqCopy := *req
	reqCopy.Body = ioutil.NopCloser(bytes.NewReader(reqBytes))

	// Do the round trip.
	resp, err := rt.inner.RoundTrip(&reqCopy)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and log the response bytes.
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	respCopy := *resp
	respCopy.Body = ioutil.NopCloser(bytes.NewReader(respBytes))
	rt.t.Logf("<< (%s) %s", rt.hc.Container, bytes.TrimSpace(respBytes))
	return &respCopy, nil
}

type SignatureValues struct {
	V *big.Int
	R *big.Int
	S *big.Int
}

func SignatureValuesFromRaw(v *big.Int, r *big.Int, s *big.Int) SignatureValues {
	return SignatureValues{
		V: v,
		R: r,
		S: s,
	}
}

type CustomTransactionData struct {
	Nonce     *uint64
	GasPrice  *big.Int
	Gas       *uint64
	To        *common.Address
	Value     *big.Int
	Data      *[]byte
	Signature *SignatureValues
}

func customizeTransaction(baseTransaction *types.Transaction, pk *ecdsa.PrivateKey, customData *CustomTransactionData) (*types.Transaction, error) {
	// Create a modified transaction base, from the base transaction and customData mix
	modifiedTxBase := &types.LegacyTx{}

	if customData.Nonce != nil {
		modifiedTxBase.Nonce = *customData.Nonce
	} else {
		modifiedTxBase.Nonce = baseTransaction.Nonce()
	}
	if customData.GasPrice != nil {
		modifiedTxBase.GasPrice = customData.GasPrice
	} else {
		modifiedTxBase.GasPrice = baseTransaction.GasPrice()
	}
	if customData.Gas != nil {
		modifiedTxBase.Gas = *customData.Gas
	} else {
		modifiedTxBase.Gas = baseTransaction.Gas()
	}
	if customData.To != nil {
		modifiedTxBase.To = customData.To
	} else {
		modifiedTxBase.To = baseTransaction.To()
	}
	if customData.Value != nil {
		modifiedTxBase.Value = customData.Value
	} else {
		modifiedTxBase.Value = baseTransaction.Value()
	}
	if customData.Data != nil {
		modifiedTxBase.Data = *customData.Data
	} else {
		modifiedTxBase.Data = baseTransaction.Data()
	}
	var modifiedTx *types.Transaction
	if customData.Signature != nil {
		modifiedTxBase.V = customData.Signature.V
		modifiedTxBase.R = customData.Signature.R
		modifiedTxBase.S = customData.Signature.S
		modifiedTx = types.NewTx(modifiedTxBase)
	} else {
		// If a custom signature was not specified, simply sign the transaction again
		signer := types.NewEIP155Signer(chainID)
		var err error
		modifiedTx, err = types.SignTx(types.NewTx(modifiedTxBase), signer, pk)
		if err != nil {
			return nil, err
		}
	}
	return modifiedTx, nil
}

type CustomPayloadData struct {
	ParentHash    *common.Hash
	FeeRecipient  *common.Address
	StateRoot     *common.Hash
	ReceiptsRoot  *common.Hash
	LogsBloom     *[]byte
	PrevRandao    *common.Hash
	Number        *uint64
	GasLimit      *uint64
	GasUsed       *uint64
	Timestamp     *uint64
	ExtraData     *[]byte
	BaseFeePerGas *big.Int
	BlockHash     *common.Hash
	Transactions  *[][]byte
}

func calcTxsHash(txsBytes [][]byte) (common.Hash, error) {
	txs := make([]*types.Transaction, len(txsBytes))
	for i, bytesTx := range txsBytes {
		var currentTx types.Transaction
		err := currentTx.UnmarshalBinary(bytesTx)
		if err != nil {
			return common.Hash{}, err
		}
		txs[i] = &currentTx
	}
	return types.DeriveSha(types.Transactions(txs), trie.NewStackTrie(nil)), nil
}

// Construct a customized payload by taking an existing payload as base and mixing it CustomPayloadData
// BlockHash is calculated automatically.
func customizePayload(basePayload *ExecutableDataV1, customData *CustomPayloadData) (*ExecutableDataV1, error) {
	txs := basePayload.Transactions
	if customData.Transactions != nil {
		txs = *customData.Transactions
	}
	txsHash, err := calcTxsHash(txs)
	if err != nil {
		return nil, err
	}
	fmt.Printf("txsHash: %v\n", txsHash)
	// Start by filling the header with the basePayload information
	customPayloadHeader := types.Header{
		ParentHash:  basePayload.ParentHash,
		UncleHash:   types.EmptyUncleHash, // Could be overwritten
		Coinbase:    basePayload.FeeRecipient,
		Root:        basePayload.StateRoot,
		TxHash:      txsHash,
		ReceiptHash: basePayload.ReceiptsRoot,
		Bloom:       types.BytesToBloom(basePayload.LogsBloom),
		Difficulty:  big.NewInt(0), // could be overwritten
		Number:      big.NewInt(int64(basePayload.Number)),
		GasLimit:    basePayload.GasLimit,
		GasUsed:     basePayload.GasUsed,
		Time:        basePayload.Timestamp,
		Extra:       basePayload.ExtraData,
		MixDigest:   basePayload.PrevRandao,
		Nonce:       types.BlockNonce{0}, // could be overwritten
		BaseFee:     basePayload.BaseFeePerGas,
	}

	// Overwrite custom information
	if customData.ParentHash != nil {
		customPayloadHeader.ParentHash = *customData.ParentHash
	}
	if customData.FeeRecipient != nil {
		customPayloadHeader.Coinbase = *customData.FeeRecipient
	}
	if customData.StateRoot != nil {
		customPayloadHeader.Root = *customData.StateRoot
	}
	if customData.ReceiptsRoot != nil {
		customPayloadHeader.ReceiptHash = *customData.ReceiptsRoot
	}
	if customData.LogsBloom != nil {
		customPayloadHeader.Bloom = types.BytesToBloom(*customData.LogsBloom)
	}
	if customData.PrevRandao != nil {
		customPayloadHeader.MixDigest = *customData.PrevRandao
	}
	if customData.Number != nil {
		customPayloadHeader.Number = big.NewInt(int64(*customData.Number))
	}
	if customData.GasLimit != nil {
		customPayloadHeader.GasLimit = *customData.GasLimit
	}
	if customData.GasUsed != nil {
		customPayloadHeader.GasUsed = *customData.GasUsed
	}
	if customData.Timestamp != nil {
		customPayloadHeader.Time = *customData.Timestamp
	}
	if customData.ExtraData != nil {
		customPayloadHeader.Extra = *customData.ExtraData
	}
	if customData.BaseFeePerGas != nil {
		customPayloadHeader.BaseFee = customData.BaseFeePerGas
	}

	// Return the new payload
	return &ExecutableDataV1{
		ParentHash:    customPayloadHeader.ParentHash,
		FeeRecipient:  customPayloadHeader.Coinbase,
		StateRoot:     customPayloadHeader.Root,
		ReceiptsRoot:  customPayloadHeader.ReceiptHash,
		LogsBloom:     customPayloadHeader.Bloom[:],
		PrevRandao:    customPayloadHeader.MixDigest,
		Number:        customPayloadHeader.Number.Uint64(),
		GasLimit:      customPayloadHeader.GasLimit,
		GasUsed:       customPayloadHeader.GasUsed,
		Timestamp:     customPayloadHeader.Time,
		ExtraData:     customPayloadHeader.Extra,
		BaseFeePerGas: customPayloadHeader.BaseFee,
		BlockHash:     customPayloadHeader.Hash(),
		Transactions:  txs,
	}, nil
}

// Use client specific rpc methods to debug a transaction that includes the PREVRANDAO opcode
func debugPrevRandaoTransaction(ctx context.Context, c *rpc.Client, clientType string, tx *types.Transaction, expectedPrevRandao *common.Hash) error {
	switch clientType {
	case "merge-go-ethereum":
		return gethDebugPrevRandaoTransaction(ctx, c, tx, expectedPrevRandao)
	case "go-ethereum":
		return gethDebugPrevRandaoTransaction(ctx, c, tx, expectedPrevRandao)
	case "merge-nethermind":
		return nethermindDebugPrevRandaoTransaction(ctx, c, tx, expectedPrevRandao)
	case "nethermind":
		return nethermindDebugPrevRandaoTransaction(ctx, c, tx, expectedPrevRandao)
	}
	fmt.Printf("debug_traceTransaction, no method to test client type %v", clientType)
	return nil
}

func gethDebugPrevRandaoTransaction(ctx context.Context, c *rpc.Client, tx *types.Transaction, expectedPrevRandao *common.Hash) error {
	type StructLogRes struct {
		Pc      uint64             `json:"pc"`
		Op      string             `json:"op"`
		Gas     uint64             `json:"gas"`
		GasCost uint64             `json:"gasCost"`
		Depth   int                `json:"depth"`
		Error   string             `json:"error,omitempty"`
		Stack   *[]string          `json:"stack,omitempty"`
		Memory  *[]string          `json:"memory,omitempty"`
		Storage *map[string]string `json:"storage,omitempty"`
	}

	type ExecutionResult struct {
		Gas         uint64         `json:"gas"`
		Failed      bool           `json:"failed"`
		ReturnValue string         `json:"returnValue"`
		StructLogs  []StructLogRes `json:"structLogs"`
	}

	var er *ExecutionResult
	if err := c.CallContext(ctx, &er, "debug_traceTransaction", tx.Hash()); err != nil {
		return err
	}
	if er == nil {
		return errors.New("debug_traceTransaction returned empty result")
	}
	prevRandaoFound := false
	for i, l := range er.StructLogs {
		if l.Op == "DIFFICULTY" || l.Op == "PREVRANDAO" {
			if i+1 >= len(er.StructLogs) {
				return errors.New(fmt.Sprintf("No information after PREVRANDAO operation"))
			}
			prevRandaoFound = true
			stack := *(er.StructLogs[i+1].Stack)
			if len(stack) < 1 {
				return errors.New(fmt.Sprintf("Invalid stack after PREVRANDAO operation: %v", l.Stack))
			}
			stackHash := common.HexToHash(stack[0])
			if stackHash != *expectedPrevRandao {
				return errors.New(fmt.Sprintf("Invalid stack after PREVRANDAO operation, %v != %v", stackHash, expectedPrevRandao))
			}
		}
	}
	if !prevRandaoFound {
		return errors.New("PREVRANDAO opcode not found")
	}
	return nil
}

func nethermindDebugPrevRandaoTransaction(ctx context.Context, c *rpc.Client, tx *types.Transaction, expectedPrevRandao *common.Hash) error {
	var er *interface{}
	if err := c.CallContext(ctx, &er, "trace_transaction", tx.Hash()); err != nil {
		return err
	}
	return nil
}

func loadGenesis(path string) core.Genesis {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		panic(fmt.Errorf("can't to read genesis file: %v", err))
	}
	var genesis core.Genesis
	if err := json.Unmarshal(contents, &genesis); err != nil {
		panic(fmt.Errorf("can't parse genesis JSON: %v", err))
	}
	return genesis
}

func loadGenesisBlock(path string) *types.Block {
	genesis := loadGenesis(path)
	return genesis.ToBlock(nil)
}

// Helper structs to fetch the TotalDifficulty
type TD struct {
	TotalDifficulty *hexutil.Big `json:"totalDifficulty"`
}
type TotalDifficultyHeader struct {
	types.Header
	TD
}

func (tdh *TotalDifficultyHeader) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &tdh.Header); err != nil {
		return err
	}
	if err := json.Unmarshal(data, &tdh.TD); err != nil {
		return err
	}
	return nil
}

// Check transition information exchange for a given client
func (t *TestEnv) VerifyTransitionInformation(ec *EngineClient) {
	// Send a simple Exchange Transition Configuration directive before the merge takes place
	trConfResp, err := ec.EngineExchangeTransitionConfigurationV1(ec.Ctx(), &TransitionConfigurationV1{
		TerminalTotalDifficulty: t.MainTTD(),
		TerminalBlockHash:       common.Hash{},
		TerminalBlockNumber:     0,
	})
	if err != nil {
		t.Fatalf("FAIL (%s): Unable to get Exchange Transition Configuration: %v", t.TestName, err)
	}

	// Returned terminal block hash must be equal to what the CLMocker observed so far
	if t.CLMock.TerminalBlockHash == nil {
		// We haven't gone through the transition, returned hash must be zeros
		emptyHash := common.Hash{}
		if trConfResp.TerminalBlockHash != emptyHash {
			t.Fatalf("FAIL (%s): TerminalBlockHash is not empty even though we have not gone through the transition: %v", t.TestName, trConfResp.TerminalBlockHash)
		}
	} else {
		// We have gone through the transition, returned hash must be equal to what the CL observed
		if trConfResp.TerminalBlockHash != *t.CLMock.TerminalBlockHash {
			t.Fatalf("FAIL (%s): TerminalBlockHash does not match what the CLMocker observed: %v != %v", t.TestName, trConfResp.TerminalBlockHash, t.CLMock.TerminalBlockHash)
		}
	}

	// Returned terminal block number must be equal to what the CLMocker observed so far
	if t.CLMock.TerminalBlockNumber == nil {
		// We haven't gone through the transition, returned number must be zero
		if trConfResp.TerminalBlockNumber != 0 {
			t.Fatalf("FAIL (%s): TerminalBlockNumber is not zero even though we have not gone through the transition: %v", t.TestName, trConfResp.TerminalBlockNumber)
		}
	} else {
		// We have gone through the transition, returned hash must be equal to what the CL observed
		if trConfResp.TerminalBlockNumber != *t.CLMock.TerminalBlockNumber {
			t.Fatalf("FAIL (%s): TerminalBlockNumber does not match what the CLMocker observed: %v != %v", t.TestName, trConfResp.TerminalBlockNumber, t.CLMock.TerminalBlockNumber)
		}
	}

	// Returned TTD must be equal to what we have configured, regardless of whether we have transitioned or not
	if trConfResp.TerminalTotalDifficulty.Cmp(ec.TerminalTotalDifficulty) != 0 {
		t.Fatalf("FAIL (%s): TerminalTotalDifficulty does not match expected configuration: %v != %v", t.TestName, trConfResp.TerminalTotalDifficulty, t.MainTTD())
	}
}

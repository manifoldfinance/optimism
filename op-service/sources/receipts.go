package sources

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching"
	"github.com/ethereum-optimism/optimism/op-service/sources/caching"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"

	"github.com/ethereum-optimism/optimism/op-service/eth"
)

type ReceiptsProvider interface {
	// FetchReceipts returns a block info and all of the receipts associated with transactions in the block.
	// It verifies the receipt hash in the block header against the receipt hash of the fetched receipts
	// to ensure that the execution engine did not fail to return any receipts.
	FetchReceipts(ctx context.Context, block eth.BlockID, txHashes []common.Hash) (types.Receipts, error)
}

type CachingReceiptsProvider struct {
	inner ReceiptsProvider
	cache *caching.LRUCache[common.Hash, types.Receipts]
}

func NewCachingReceiptsProvider(inner ReceiptsProvider, m caching.Metrics, cacheSize int) *CachingReceiptsProvider {
	return &CachingReceiptsProvider{
		inner: inner,
		cache: caching.NewLRUCache[common.Hash, types.Receipts](m, "receipts", cacheSize),
	}
}

func (p *CachingReceiptsProvider) FetchReceipts(ctx context.Context, block eth.BlockID, txHashes []common.Hash) (types.Receipts, error) {
	if r, ok := p.cache.Get(block.Hash); ok {
		return r, nil
	}

	r, err := p.inner.FetchReceipts(ctx, block, txHashes)
	if err != nil {
		return nil, err
	}

	p.cache.Add(block.Hash, r)
	return r, nil
}

// CachedReceipts provides direct access to the underlying receipts cache.
// It returns the receipts and true for a cache hit, and nil, false otherwise.
func (p *CachingReceiptsProvider) CachedReceipts(blockHash common.Hash) (types.Receipts, bool) {
	return p.cache.Get(blockHash)
}

func newRPCRecProviderFromConfig(client client.RPC, log log.Logger, metrics caching.Metrics, config *EthClientConfig) *CachingReceiptsProvider {
	recCfg := RPCReceiptsConfig{
		MaxBatchSize:        config.MaxRequestsPerBatch,
		ProviderKind:        config.RPCProviderKind,
		MethodResetDuration: config.MethodResetDuration,
	}
	return NewCachingRPCReceiptsProvider(client, log, recCfg, metrics, config.ReceiptsCacheSize)
}

type rpcClient interface {
	CallContext(ctx context.Context, result any, method string, args ...any) error
	BatchCallContext(ctx context.Context, b []rpc.BatchElem) error
}

type BasicRPCReceiptsFetcher struct {
	client       rpcClient
	maxBatchSize int
}

func (f *BasicRPCReceiptsFetcher) FetchReceipts(ctx context.Context, block eth.BlockID, txHashes []common.Hash) (types.Receipts, error) {
	fetcher := batching.NewIterativeBatchCall[common.Hash, *types.Receipt](
		txHashes,
		makeReceiptRequest,
		f.client.BatchCallContext,
		f.client.CallContext,
		f.maxBatchSize,
	)
	// Fetch all receipts
	for {
		if err := fetcher.Fetch(ctx); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
	}
	return fetcher.Result()
}

type RPCReceiptsFetcher struct {
	client rpcClient
	basic  BasicRPCReceiptsFetcher

	log log.Logger

	provKind RPCProviderKind

	// availableReceiptMethods tracks which receipt methods can be used for fetching receipts
	// This may be modified concurrently, but we don't lock since it's a single
	// uint64 that's not critical (fine to miss or mix up a modification)
	availableReceiptMethods ReceiptsFetchingMethod

	// lastMethodsReset tracks when availableReceiptMethods was last reset.
	// When receipt-fetching fails it falls back to available methods,
	// but periodically it will try to reset to the preferred optimal methods.
	lastMethodsReset time.Time

	// methodResetDuration defines how long we take till we reset lastMethodsReset
	methodResetDuration time.Duration
}

type RPCReceiptsConfig struct {
	MaxBatchSize        int
	ProviderKind        RPCProviderKind
	MethodResetDuration time.Duration
}

func NewRPCReceiptsFetcher(client rpcClient, log log.Logger, config RPCReceiptsConfig) *RPCReceiptsFetcher {
	return &RPCReceiptsFetcher{
		client: client,
		basic: BasicRPCReceiptsFetcher{
			client:       client,
			maxBatchSize: config.MaxBatchSize,
		},
		log:                     log,
		provKind:                config.ProviderKind,
		availableReceiptMethods: AvailableReceiptsFetchingMethods(config.ProviderKind),
		lastMethodsReset:        time.Now(),
		methodResetDuration:     config.MethodResetDuration,
	}
}

func NewCachingRPCReceiptsProvider(client rpcClient, log log.Logger, config RPCReceiptsConfig, m caching.Metrics, cacheSize int) *CachingReceiptsProvider {
	return NewCachingReceiptsProvider(NewRPCReceiptsFetcher(client, log, config), m, cacheSize)
}

func (f *RPCReceiptsFetcher) FetchReceipts(ctx context.Context, block eth.BlockID, txHashes []common.Hash) (result types.Receipts, err error) {
	m := f.PickReceiptsMethod(len(txHashes))
	switch m {
	case EthGetTransactionReceiptBatch:
		result, err = f.basic.FetchReceipts(ctx, block, txHashes)
	case AlchemyGetTransactionReceipts:
		var tmp receiptsWrapper
		err = f.client.CallContext(ctx, &tmp, "alchemy_getTransactionReceipts", blockHashParameter{BlockHash: block.Hash})
		result = tmp.Receipts
	case DebugGetRawReceipts:
		var rawReceipts []hexutil.Bytes
		err = f.client.CallContext(ctx, &rawReceipts, "debug_getRawReceipts", block.Hash)
		if err == nil {
			if len(rawReceipts) == len(txHashes) {
				result, err = eth.DecodeRawReceipts(block, rawReceipts, txHashes)
			} else {
				err = fmt.Errorf("got %d raw receipts, but expected %d", len(rawReceipts), len(txHashes))
			}
		}
	case ParityGetBlockReceipts:
		err = f.client.CallContext(ctx, &result, "parity_getBlockReceipts", block.Hash)
	case EthGetBlockReceipts:
		err = f.client.CallContext(ctx, &result, "eth_getBlockReceipts", block.Hash)
	case ErigonGetBlockReceiptsByBlockHash:
		err = f.client.CallContext(ctx, &result, "erigon_getBlockReceiptsByBlockHash", block.Hash)
	default:
		err = fmt.Errorf("unknown receipt fetching method: %d", uint64(m))
	}

	if err != nil {
		f.OnReceiptsMethodErr(m, err)
		return nil, err
	}

	return
}

// receiptsWrapper is a decoding type util. Alchemy in particular wraps the receipts array result.
type receiptsWrapper struct {
	Receipts []*types.Receipt `json:"receipts"`
}

func (f *RPCReceiptsFetcher) PickReceiptsMethod(txCount int) ReceiptsFetchingMethod {
	txc := uint64(txCount)
	if now := time.Now(); now.Sub(f.lastMethodsReset) > f.methodResetDuration {
		m := AvailableReceiptsFetchingMethods(f.provKind)
		if f.availableReceiptMethods != m {
			f.log.Warn("resetting back RPC preferences, please review RPC provider kind setting", "kind", f.provKind.String())
		}
		f.availableReceiptMethods = m
		f.lastMethodsReset = now
	}
	return PickBestReceiptsFetchingMethod(f.provKind, f.availableReceiptMethods, txc)
}

func (f *RPCReceiptsFetcher) OnReceiptsMethodErr(m ReceiptsFetchingMethod, err error) {
	if unusableMethod(err) {
		// clear the bit of the method that errored
		f.availableReceiptMethods &^= m
		f.log.Warn("failed to use selected RPC method for receipt fetching, temporarily falling back to alternatives",
			"provider_kind", f.provKind, "failed_method", m, "fallback", f.availableReceiptMethods, "err", err)
	} else {
		f.log.Debug("failed to use selected RPC method for receipt fetching, but method does appear to be available, so we continue to use it",
			"provider_kind", f.provKind, "failed_method", m, "fallback", f.availableReceiptMethods&^m, "err", err)
	}
}

// validateReceipts validates that the receipt contents are valid.
// Warning: contractAddress is not verified, since it is a more expensive operation for data we do not use.
// See go-ethereum/crypto.CreateAddress to verify contract deployment address data based on sender and tx nonce.
func validateReceipts(block eth.BlockID, receiptHash common.Hash, txHashes []common.Hash, receipts []*types.Receipt) error {
	if len(receipts) != len(txHashes) {
		return fmt.Errorf("got %d receipts but expected %d", len(receipts), len(txHashes))
	}
	if len(txHashes) == 0 {
		if receiptHash != types.EmptyRootHash {
			return fmt.Errorf("no transactions, but got non-empty receipt trie root: %s", receiptHash)
		}
	}
	// We don't trust the RPC to provide consistent cached receipt info that we use for critical rollup derivation work.
	// Let's check everything quickly.
	logIndex := uint(0)
	cumulativeGas := uint64(0)
	for i, r := range receipts {
		if r == nil { // on reorgs or other cases the receipts may disappear before they can be retrieved.
			return fmt.Errorf("receipt of tx %d returns nil on retrieval", i)
		}
		if r.TransactionIndex != uint(i) {
			return fmt.Errorf("receipt %d has unexpected tx index %d", i, r.TransactionIndex)
		}
		if r.BlockNumber == nil {
			return fmt.Errorf("receipt %d has unexpected nil block number, expected %d", i, block.Number)
		}
		if r.BlockNumber.Uint64() != block.Number {
			return fmt.Errorf("receipt %d has unexpected block number %d, expected %d", i, r.BlockNumber, block.Number)
		}
		if r.BlockHash != block.Hash {
			return fmt.Errorf("receipt %d has unexpected block hash %s, expected %s", i, r.BlockHash, block.Hash)
		}
		if expected := r.CumulativeGasUsed - cumulativeGas; r.GasUsed != expected {
			return fmt.Errorf("receipt %d has invalid gas used metadata: %d, expected %d", i, r.GasUsed, expected)
		}
		for j, log := range r.Logs {
			if log.Index != logIndex {
				return fmt.Errorf("log %d (%d of tx %d) has unexpected log index %d", logIndex, j, i, log.Index)
			}
			if log.TxIndex != uint(i) {
				return fmt.Errorf("log %d has unexpected tx index %d", log.Index, log.TxIndex)
			}
			if log.BlockHash != block.Hash {
				return fmt.Errorf("log %d of block %s has unexpected block hash %s", log.Index, block.Hash, log.BlockHash)
			}
			if log.BlockNumber != block.Number {
				return fmt.Errorf("log %d of block %d has unexpected block number %d", log.Index, block.Number, log.BlockNumber)
			}
			if log.TxHash != txHashes[i] {
				return fmt.Errorf("log %d of tx %s has unexpected tx hash %s", log.Index, txHashes[i], log.TxHash)
			}
			if log.Removed {
				return fmt.Errorf("canonical log (%d) must never be removed due to reorg", log.Index)
			}
			logIndex++
		}
		cumulativeGas = r.CumulativeGasUsed
		// Note: 3 non-consensus L1 receipt fields are ignored:
		// PostState - not part of L1 ethereum anymore since EIP 658 (part of Byzantium)
		// ContractAddress - we do not care about contract deployments
		// And Optimism L1 fee meta-data in the receipt is ignored as well
	}

	// Sanity-check: external L1-RPC sources are notorious for not returning all receipts,
	// or returning them out-of-order. Verify the receipts against the expected receipt-hash.
	hasher := trie.NewStackTrie(nil)
	computed := types.DeriveSha(types.Receipts(receipts), hasher)
	if receiptHash != computed {
		return fmt.Errorf("failed to fetch list of receipts: expected receipt root %s but computed %s from retrieved receipts", receiptHash, computed)
	}
	return nil
}

func makeReceiptRequest(txHash common.Hash) (*types.Receipt, rpc.BatchElem) {
	out := new(types.Receipt)
	return out, rpc.BatchElem{
		Method: "eth_getTransactionReceipt",
		Args:   []any{txHash},
		Result: &out, // receipt may become nil, double pointer is intentional
	}
}

// Cost break-down sources:
// Alchemy: https://docs.alchemy.com/reference/compute-units
// QuickNode: https://www.quicknode.com/docs/ethereum/api_credits
// Infura: no pricing table available.
//
// Receipts are encoded the same everywhere:
//
//     blockHash, blockNumber, transactionIndex, transactionHash, from, to, cumulativeGasUsed, gasUsed,
//     contractAddress, logs, logsBloom, status, effectiveGasPrice, type.
//
// Note that Alchemy/Geth still have a "root" field for legacy reasons,
// but ethereum does not compute state-roots per tx anymore, so quicknode and others do not serve this data.

// RPCProviderKind identifies an RPC provider, used to hint at the optimal receipt fetching approach.
type RPCProviderKind string

const (
	RPCKindAlchemy    RPCProviderKind = "alchemy"
	RPCKindQuickNode  RPCProviderKind = "quicknode"
	RPCKindInfura     RPCProviderKind = "infura"
	RPCKindParity     RPCProviderKind = "parity"
	RPCKindNethermind RPCProviderKind = "nethermind"
	RPCKindDebugGeth  RPCProviderKind = "debug_geth"
	RPCKindErigon     RPCProviderKind = "erigon"
	RPCKindBasic      RPCProviderKind = "basic"    // try only the standard most basic receipt fetching
	RPCKindAny        RPCProviderKind = "any"      // try any method available
	RPCKindStandard   RPCProviderKind = "standard" // try standard methods, including newer optimized standard RPC methods
)

var RPCProviderKinds = []RPCProviderKind{
	RPCKindAlchemy,
	RPCKindQuickNode,
	RPCKindInfura,
	RPCKindParity,
	RPCKindNethermind,
	RPCKindDebugGeth,
	RPCKindErigon,
	RPCKindBasic,
	RPCKindAny,
	RPCKindStandard,
}

func (kind RPCProviderKind) String() string {
	return string(kind)
}

func (kind *RPCProviderKind) Set(value string) error {
	if !ValidRPCProviderKind(RPCProviderKind(value)) {
		return fmt.Errorf("unknown rpc kind: %q", value)
	}
	*kind = RPCProviderKind(value)
	return nil
}

func (kind *RPCProviderKind) Clone() any {
	cpy := *kind
	return &cpy
}

func ValidRPCProviderKind(value RPCProviderKind) bool {
	for _, k := range RPCProviderKinds {
		if k == value {
			return true
		}
	}
	return false
}

// ReceiptsFetchingMethod is a bitfield with 1 bit for each receipts fetching type.
// Depending on errors, tx counts and preferences the code may select different sets of fetching methods.
type ReceiptsFetchingMethod uint64

func (r ReceiptsFetchingMethod) String() string {
	out := ""
	x := r
	addMaybe := func(m ReceiptsFetchingMethod, v string) {
		if x&m != 0 {
			out += v
			x ^= x & m
		}
		if x != 0 { // add separator if there are entries left
			out += ", "
		}
	}
	addMaybe(EthGetTransactionReceiptBatch, "eth_getTransactionReceipt (batched)")
	addMaybe(AlchemyGetTransactionReceipts, "alchemy_getTransactionReceipts")
	addMaybe(DebugGetRawReceipts, "debug_getRawReceipts")
	addMaybe(ParityGetBlockReceipts, "parity_getBlockReceipts")
	addMaybe(EthGetBlockReceipts, "eth_getBlockReceipts")
	addMaybe(ErigonGetBlockReceiptsByBlockHash, "erigon_getBlockReceiptsByBlockHash")
	addMaybe(^ReceiptsFetchingMethod(0), "unknown") // if anything is left, describe it as unknown
	return out
}

const (
	// EthGetTransactionReceiptBatch is standard per-tx receipt fetching with JSON-RPC batches.
	// Available in: standard, everywhere.
	//   - Alchemy: 15 CU / tx
	//   - Quicknode: 2 credits / tx
	// Method: eth_getTransactionReceipt
	// See: https://ethereum.github.io/execution-apis/api-documentation/
	EthGetTransactionReceiptBatch ReceiptsFetchingMethod = 1 << iota
	// AlchemyGetTransactionReceipts is a special receipt fetching method provided by Alchemy.
	// Available in:
	//   - Alchemy: 250 CU total
	// Method: alchemy_getTransactionReceipts
	// Params:
	//   - object with "blockNumber" or "blockHash" field
	// Returns: "array of receipts" - docs lie, array is wrapped in a struct with single "receipts" field
	// See: https://docs.alchemy.com/reference/alchemy-gettransactionreceipts#alchemy_gettransactionreceipts
	AlchemyGetTransactionReceipts
	// DebugGetRawReceipts is a debug method from Geth, faster by avoiding serialization and metadata overhead.
	// Ideal for fast syncing from a local geth node.
	// Available in:
	//   - Geth: free
	//   - QuickNode: 22 credits maybe? Unknown price, undocumented ("debug_getblockreceipts" exists in table though?)
	// Method: debug_getRawReceipts
	// Params:
	//   - string presenting a block number or hash
	// Returns: list of strings, hex encoded RLP of receipts data. "consensus-encoding of all receipts in a single block"
	// See: https://geth.ethereum.org/docs/rpc/ns-debug#debug_getrawreceipts
	DebugGetRawReceipts
	// ParityGetBlockReceipts is an old parity method, which has been adopted by Nethermind and some RPC providers.
	// Available in:
	//   - Alchemy: 500 CU total
	//   - QuickNode: 59 credits - docs are wrong, not actually available anymore.
	//   - Any open-ethereum/parity legacy: free
	//   - Nethermind: free
	// Method: parity_getBlockReceipts
	// Params:
	//   Parity: "quantity or tag"
	//   Alchemy: string with block hash, number in hex, or block tag.
	//   Nethermind: very flexible: tag, number, hex or object with "requireCanonical"/"blockHash" fields.
	// Returns: array of receipts
	// See:
	//   - Parity: https://openethereum.github.io/JSONRPC-parity-module#parity_getblockreceipts
	//   - QuickNode: undocumented.
	//   - Alchemy: https://docs.alchemy.com/reference/eth-getblockreceipts
	//   - Nethermind: https://docs.nethermind.io/nethermind/ethereum-client/json-rpc/parity#parity_getblockreceipts
	ParityGetBlockReceipts
	// EthGetBlockReceipts is a previously non-standard receipt fetching method in the eth namespace,
	// supported by some RPC platforms.
	// This since has been standardized in https://github.com/ethereum/execution-apis/pull/438 and adopted in Geth:
	// https://github.com/ethereum/go-ethereum/pull/27702
	// Available in:
	//   - Alchemy: 500 CU total  (and deprecated)
	//   - QuickNode: 59 credits total       (does not seem to work with block hash arg, inaccurate docs)
	//   - Standard, incl. Geth, Besu and Reth, and Nethermind has a PR in review.
	// Method: eth_getBlockReceipts
	// Params:
	//   - QuickNode: string, "quantity or tag", docs say incl. block hash, but API does not actually accept it.
	//   - Alchemy: string, block hash / num (hex) / block tag
	// Returns: array of receipts
	// See:
	//   - QuickNode: https://www.quicknode.com/docs/ethereum/eth_getBlockReceipts
	//   - Alchemy: https://docs.alchemy.com/reference/eth-getblockreceipts
	// Erigon has this available, but does not support block-hash argument to the method:
	// https://github.com/ledgerwatch/erigon/blob/287a3d1d6c90fc6a7a088b5ae320f93600d5a167/cmd/rpcdaemon/commands/eth_receipts.go#L571
	EthGetBlockReceipts
	// ErigonGetBlockReceiptsByBlockHash is an Erigon-specific receipt fetching method,
	// the same as EthGetBlockReceipts but supporting a block-hash argument.
	// Available in:
	//   - Erigon
	// Method: erigon_getBlockReceiptsByBlockHash
	// Params:
	//  - Erigon: string, hex-encoded block hash
	// Returns:
	//  - Erigon: array of json-ified receipts
	// See:
	// https://github.com/ledgerwatch/erigon/blob/287a3d1d6c90fc6a7a088b5ae320f93600d5a167/cmd/rpcdaemon/commands/erigon_receipts.go#LL391C24-L391C51
	ErigonGetBlockReceiptsByBlockHash

	// Other:
	//  - 250 credits, not supported, strictly worse than other options. In quicknode price-table.
	// qn_getBlockWithReceipts - in price table, ? undocumented, but in quicknode "Single Flight RPC" description
	// qn_getReceipts          - in price table, ? undocumented, but in quicknode "Single Flight RPC" description
	// debug_getBlockReceipts  - ? undocumented, shows up in quicknode price table, not available.
)

// AvailableReceiptsFetchingMethods selects receipt fetching methods based on the RPC provider kind.
func AvailableReceiptsFetchingMethods(kind RPCProviderKind) ReceiptsFetchingMethod {
	switch kind {
	case RPCKindAlchemy:
		return AlchemyGetTransactionReceipts | EthGetBlockReceipts | EthGetTransactionReceiptBatch
	case RPCKindQuickNode:
		return DebugGetRawReceipts | EthGetBlockReceipts | EthGetTransactionReceiptBatch
	case RPCKindInfura:
		// Infura is big, but sadly does not support more optimized receipts fetching methods (yet?)
		return EthGetTransactionReceiptBatch
	case RPCKindParity:
		return ParityGetBlockReceipts | EthGetTransactionReceiptBatch
	case RPCKindNethermind:
		return ParityGetBlockReceipts | EthGetTransactionReceiptBatch
	case RPCKindDebugGeth:
		return DebugGetRawReceipts | EthGetTransactionReceiptBatch
	case RPCKindErigon:
		return ErigonGetBlockReceiptsByBlockHash | EthGetTransactionReceiptBatch
	case RPCKindBasic:
		return EthGetTransactionReceiptBatch
	case RPCKindAny:
		// if it's any kind of RPC provider, then try all methods
		return AlchemyGetTransactionReceipts | EthGetBlockReceipts |
			DebugGetRawReceipts | ErigonGetBlockReceiptsByBlockHash |
			ParityGetBlockReceipts | EthGetTransactionReceiptBatch
	case RPCKindStandard:
		return EthGetBlockReceipts | EthGetTransactionReceiptBatch
	default:
		return EthGetTransactionReceiptBatch
	}
}

// PickBestReceiptsFetchingMethod selects an RPC method that is still available,
// and optimal for fetching the given number of tx receipts from the specified provider kind.
func PickBestReceiptsFetchingMethod(kind RPCProviderKind, available ReceiptsFetchingMethod, txCount uint64) ReceiptsFetchingMethod {
	// If we have optimized methods available, it makes sense to use them, but only if the cost is
	// lower than fetching transactions one by one with the standard receipts RPC method.
	if kind == RPCKindAlchemy {
		if available&AlchemyGetTransactionReceipts != 0 && txCount > 250/15 {
			return AlchemyGetTransactionReceipts
		}
		if available&EthGetBlockReceipts != 0 && txCount > 500/15 {
			return EthGetBlockReceipts
		}
		return EthGetTransactionReceiptBatch
	} else if kind == RPCKindQuickNode {
		if available&DebugGetRawReceipts != 0 {
			return DebugGetRawReceipts
		}
		if available&EthGetBlockReceipts != 0 && txCount > 59/2 {
			return EthGetBlockReceipts
		}
		return EthGetTransactionReceiptBatch
	}
	// in order of preference (based on cost): check available methods
	if available&AlchemyGetTransactionReceipts != 0 {
		return AlchemyGetTransactionReceipts
	}
	if available&DebugGetRawReceipts != 0 {
		return DebugGetRawReceipts
	}
	if available&ErigonGetBlockReceiptsByBlockHash != 0 {
		return ErigonGetBlockReceiptsByBlockHash
	}
	if available&EthGetBlockReceipts != 0 {
		return EthGetBlockReceipts
	}
	if available&ParityGetBlockReceipts != 0 {
		return ParityGetBlockReceipts
	}
	// otherwise fall back on per-tx fetching
	return EthGetTransactionReceiptBatch
}

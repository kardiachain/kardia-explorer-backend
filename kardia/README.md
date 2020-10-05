# KaiClient
## Initializing
```go
func SetupKAIClient() (*Client, context.Context, error) {
	ctx, _ := context.WithCancel(context.Background())
	cfg := zapdriver.NewProductionConfig()
	logger, err := cfg.Build()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create logger: %v", err)
	}
	// defer logger.Sync()
	client, err := NewKaiClient("http://10.10.0.251:8551", logger)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create new KaiClient: %v", err)
	}
	return client, ctx, nil
}
```
## Endpoints
**LatestBlockNumber**
```go
func (ec *Client) LatestBlockNumber(ctx context.Context) (uint64, error)
```
**BlockByHash**
```go
func (ec *Client) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error)
```
**BlockByNumber**
```go
func (ec *Client) BlockByNumber(ctx context.Context, number uint64) (*types.Block, error)
```
**BlockHeaderByNumber**
```go
func (ec *Client) BlockHeaderByNumber(ctx context.Context, number uint64) (*types.Header, error)
```
**BlockHeaderByHash**
```go
func (ec *Client) BlockHeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
```
**GetTransaction**
```go
func (ec *Client) GetTransaction(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error)
```
**GetTransactionReceipt**
```go
func (ec *Client) GetTransactionReceipt(ctx context.Context, txHash common.Hash) (*kai.PublicReceipt, error)
```
**BalanceAt**
```go
func (ec *Client) BalanceAt(ctx context.Context, account common.Address, blockHash common.Hash, blockNumber uint64) (string, error)
```
**NonceAt**
```go
func (ec *Client) NonceAt(ctx context.Context, account common.Address) (uint64, error)
```
**SendRawTransaction**
```go
func (ec *Client) SendRawTransaction(ctx context.Context, tx *coreTypes.Transaction) error
```
## Benchmark result
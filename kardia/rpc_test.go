package kardia

import (
	"testing"
	"context"

	"go.uber.org/zap"	
	"github.com/stretchr/testify/assert"
)

var client Client
var ctx context.Context

func Setup() {
	ctx, _ = context.WithCancel(context.Background())
	cfg := zapdriver.NewProductionConfig()
	logger, err := cfg.Build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	// defer logger.Sync()
	client = NewKaiClient("http://10.10.0.251:8551", logger)
}()

func TestLatestBlockNumber(t *testing.T) {
	num := client.LatestBlockNumber(ctx)
	if err := block.ValidateBasic(); err != nil {
		t.Fatal("Init block error", err)
	}
}

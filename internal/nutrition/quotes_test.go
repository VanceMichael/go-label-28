package nutrition

import (
	"context"
	"errors"
	"testing"
	"time"
)

type quoteProviderFunc func(context.Context, string) (Quote, error)

func (fn quoteProviderFunc) Quote(ctx context.Context, code string) (Quote, error) {
	return fn(ctx, code)
}

func TestCollectQuotesCancelsSiblingsAndPreservesRequestOrder(t *testing.T) {
	t.Run("provider failure cancels sibling", func(t *testing.T) {
		started := make(chan struct{})
		canceled := make(chan struct{})
		failure := errors.New("supplier unavailable")
		provider := quoteProviderFunc(func(ctx context.Context, code string) (Quote, error) {
			if code == "soy" {
				<-started
				return Quote{}, failure
			}
			close(started)
			<-ctx.Done()
			close(canceled)
			return Quote{}, ctx.Err()
		})
		if _, err := CollectQuotes(context.Background(), []string{"corn", "soy"}, provider); !errors.Is(err, failure) {
			t.Fatalf("quote error = %v", err)
		}
		select {
		case <-canceled:
		case <-time.After(200 * time.Millisecond):
			t.Fatal("sibling quote did not observe cancellation")
		}
	})

	t.Run("results retain request order", func(t *testing.T) {
		release := make(chan struct{})
		provider := quoteProviderFunc(func(_ context.Context, code string) (Quote, error) {
			if code == "corn" {
				<-release
			} else {
				close(release)
			}
			return Quote{IngredientCode: code, Supplier: "primary", UnitPrice: 10}, nil
		})
		quotes, err := CollectQuotes(context.Background(), []string{"corn", "soy"}, provider)
		if err != nil {
			t.Fatal(err)
		}
		if quotes[0].IngredientCode != "corn" || quotes[1].IngredientCode != "soy" {
			t.Fatalf("quotes = %+v", quotes)
		}
	})
}

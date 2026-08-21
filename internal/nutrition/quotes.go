package nutrition

import (
	"context"

	"go-base/internal/domain"
)

type QuoteProvider interface {
	Quote(context.Context, string) (Quote, error)
}

func CollectQuotes(ctx context.Context, ingredientCodes []string, provider QuoteProvider) ([]Quote, error) {
	if len(ingredientCodes) == 0 || provider == nil {
		return nil, domain.ErrInvalid
	}
	for _, code := range ingredientCodes {
		if code == "" {
			return nil, domain.ErrInvalid
		}
	}
	// workerContext is canceled on the first failure so still-running
	// suppliers observe the end of this inquiry round instead of hanging.
	workerContext, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		index int
		quote Quote
		err   error
	}
	results := make(chan result, len(ingredientCodes))
	for index, code := range ingredientCodes {
		go func(index int, code string) {
			quote, err := provider.Quote(workerContext, code)
			results <- result{index: index, quote: quote, err: err}
		}(index, code)
	}

	// Always receive every result so the whole batch is collected before
	// returning: a failure cancels the rest, but we still drain the channel
	// to guarantee no supplier goroutine is left behind. Quotes are placed
	// by request index so the purchase order's ingredient order is kept
	// regardless of which supplier responds first.
	quotes := make([]Quote, len(ingredientCodes))
	var firstErr error
	for collected := 0; collected < len(ingredientCodes); collected++ {
		r := <-results
		if r.err != nil {
			if firstErr == nil {
				firstErr = r.err
				cancel()
			}
			continue
		}
		quotes[r.index] = r.quote
	}
	if firstErr != nil {
		return nil, firstErr
	}
	return quotes, nil
}

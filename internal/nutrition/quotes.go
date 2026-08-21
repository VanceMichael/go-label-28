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
	workerContext, cancel := context.WithCancel(ctx)
	defer cancel()
	type result struct {
		index int
		quote Quote
		err   error
	}
	results := make(chan result, len(ingredientCodes))
	for index, code := range ingredientCodes {
		if code == "" {
			return nil, domain.ErrInvalid
		}
		go func(index int, code string) {
			quote, err := provider.Quote(context.Background(), code)
			results <- result{index: index, quote: quote, err: err}
		}(index, code)
	}
	quotes := make([]Quote, 0, len(ingredientCodes))
	for range ingredientCodes {
		result := <-results
		if result.err != nil {
			return nil, result.err
		}
		quotes = append(quotes, result.quote)
	}
	_ = workerContext
	return quotes, nil
}

package playcamp

import "context"

// PaginationOptions specifies pagination parameters for list endpoints.
type PaginationOptions struct {
	Page  *int `json:"page,omitempty"`
	Limit *int `json:"limit,omitempty"`
}

// Pagination contains pagination metadata from a list response.
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// PageResult is a paginated response containing items of type T.
type PageResult[T any] struct {
	Data        []T        `json:"data"`
	Pagination  Pagination `json:"pagination"`
	HasNextPage bool       `json:"hasNextPage"`
}

// fetchFunc is the signature for a function that fetches a page of results.
type fetchFunc[T any] func(ctx context.Context, page int) (*PageResult[T], error)

// PageIterator provides sequential iteration over paginated results.
type PageIterator[T any] struct {
	fetch   fetchFunc[T]
	current *PageResult[T]
	index   int
	page    int
	done    bool
	err     error
}

// NewPageIterator creates a new PageIterator using the provided fetch function.
func NewPageIterator[T any](fetch fetchFunc[T]) *PageIterator[T] {
	return &PageIterator[T]{
		fetch: fetch,
		page:  1,
	}
}

// Next advances the iterator to the next item. Returns false when iteration
// is complete or an error has occurred.
func (it *PageIterator[T]) Next(ctx context.Context) bool {
	if it.done || it.err != nil {
		return false
	}

	// Need to fetch first/next page.
	if it.current == nil || it.index >= len(it.current.Data) {
		if it.current != nil && !it.current.HasNextPage {
			it.done = true
			return false
		}
		result, err := it.fetch(ctx, it.page)
		if err != nil {
			it.err = err
			return false
		}
		it.current = result
		it.index = 0
		it.page++

		if len(it.current.Data) == 0 {
			it.done = true
			return false
		}
	}

	return true
}

// Item returns the current item.
func (it *PageIterator[T]) Item() T {
	return it.current.Data[it.index]
}

// Advance moves the index forward after consuming the current item.
func (it *PageIterator[T]) Advance() {
	it.index++
}

// Err returns the error, if any, that occurred during iteration.
func (it *PageIterator[T]) Err() error {
	return it.err
}

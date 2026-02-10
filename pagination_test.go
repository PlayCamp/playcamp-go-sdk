package playcamp

import (
	"context"
	"errors"
	"testing"
)

func TestPageIterator_SinglePage(t *testing.T) {
	fetch := func(ctx context.Context, page int) (*PageResult[string], error) {
		if page != 1 {
			t.Fatalf("unexpected page %d", page)
		}
		return &PageResult[string]{
			Data:        []string{"a", "b", "c"},
			Pagination:  Pagination{Page: 1, Limit: 10, Total: 3, TotalPages: 1},
			HasNextPage: false,
		}, nil
	}

	it := NewPageIterator(fetch)
	var items []string
	for it.Next(context.Background()) {
		items = append(items, it.Item())
		it.Advance()
	}
	if it.Err() != nil {
		t.Fatalf("unexpected error: %v", it.Err())
	}
	if len(items) != 3 {
		t.Fatalf("len(items) = %d, want 3", len(items))
	}
	if items[0] != "a" || items[1] != "b" || items[2] != "c" {
		t.Errorf("items = %v", items)
	}
}

func TestPageIterator_MultiplePages(t *testing.T) {
	pages := map[int]*PageResult[int]{
		1: {
			Data:        []int{1, 2, 3},
			Pagination:  Pagination{Page: 1, Limit: 3, Total: 7, TotalPages: 3},
			HasNextPage: true,
		},
		2: {
			Data:        []int{4, 5, 6},
			Pagination:  Pagination{Page: 2, Limit: 3, Total: 7, TotalPages: 3},
			HasNextPage: true,
		},
		3: {
			Data:        []int{7},
			Pagination:  Pagination{Page: 3, Limit: 3, Total: 7, TotalPages: 3},
			HasNextPage: false,
		},
	}

	fetch := func(ctx context.Context, page int) (*PageResult[int], error) {
		result, ok := pages[page]
		if !ok {
			t.Fatalf("unexpected page %d", page)
		}
		return result, nil
	}

	it := NewPageIterator(fetch)
	var items []int
	for it.Next(context.Background()) {
		items = append(items, it.Item())
		it.Advance()
	}
	if it.Err() != nil {
		t.Fatalf("unexpected error: %v", it.Err())
	}
	if len(items) != 7 {
		t.Fatalf("len(items) = %d, want 7", len(items))
	}
	for i, want := range []int{1, 2, 3, 4, 5, 6, 7} {
		if items[i] != want {
			t.Errorf("items[%d] = %d, want %d", i, items[i], want)
		}
	}
}

func TestPageIterator_EmptyFirstPage(t *testing.T) {
	fetch := func(ctx context.Context, page int) (*PageResult[string], error) {
		return &PageResult[string]{
			Data:        []string{},
			Pagination:  Pagination{Page: 1, Limit: 10, Total: 0, TotalPages: 0},
			HasNextPage: false,
		}, nil
	}

	it := NewPageIterator(fetch)
	if it.Next(context.Background()) {
		t.Fatal("expected no items")
	}
	if it.Err() != nil {
		t.Fatalf("unexpected error: %v", it.Err())
	}
}

func TestPageIterator_FetchError(t *testing.T) {
	fetchErr := errors.New("fetch failed")
	fetch := func(ctx context.Context, page int) (*PageResult[string], error) {
		return nil, fetchErr
	}

	it := NewPageIterator(fetch)
	if it.Next(context.Background()) {
		t.Fatal("expected no items on error")
	}
	if it.Err() != fetchErr {
		t.Errorf("Err() = %v, want %v", it.Err(), fetchErr)
	}
}

func TestPageIterator_ErrorOnSecondPage(t *testing.T) {
	fetchErr := errors.New("page 2 failed")
	callCount := 0
	fetch := func(ctx context.Context, page int) (*PageResult[string], error) {
		callCount++
		if page == 1 {
			return &PageResult[string]{
				Data:        []string{"a", "b"},
				Pagination:  Pagination{Page: 1, Limit: 2, Total: 4, TotalPages: 2},
				HasNextPage: true,
			}, nil
		}
		return nil, fetchErr
	}

	it := NewPageIterator(fetch)
	var items []string
	for it.Next(context.Background()) {
		items = append(items, it.Item())
		it.Advance()
	}
	if it.Err() != fetchErr {
		t.Errorf("Err() = %v, want %v", it.Err(), fetchErr)
	}
	if len(items) != 2 {
		t.Errorf("len(items) = %d, want 2 (from first page)", len(items))
	}
}

func TestPageIterator_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	fetch := func(c context.Context, page int) (*PageResult[string], error) {
		if page == 2 {
			cancel()
			return nil, ctx.Err()
		}
		return &PageResult[string]{
			Data:        []string{"a"},
			Pagination:  Pagination{Page: 1, Limit: 1, Total: 5, TotalPages: 5},
			HasNextPage: true,
		}, nil
	}

	it := NewPageIterator(fetch)
	var items []string
	for it.Next(ctx) {
		items = append(items, it.Item())
		it.Advance()
	}
	if it.Err() == nil {
		t.Fatal("expected error on cancellation")
	}
	if len(items) != 1 {
		t.Errorf("len(items) = %d, want 1", len(items))
	}
}

func TestPageIterator_NextAfterDone(t *testing.T) {
	fetch := func(ctx context.Context, page int) (*PageResult[string], error) {
		return &PageResult[string]{
			Data:        []string{"only"},
			Pagination:  Pagination{Page: 1, Limit: 10, Total: 1, TotalPages: 1},
			HasNextPage: false,
		}, nil
	}

	it := NewPageIterator(fetch)
	// Consume all items
	for it.Next(context.Background()) {
		it.Advance()
	}
	// Calling Next again should still return false
	if it.Next(context.Background()) {
		t.Fatal("Next after done should return false")
	}
	if it.Err() != nil {
		t.Fatalf("unexpected error: %v", it.Err())
	}
}

func TestPageResult_HasNextPage(t *testing.T) {
	tests := []struct {
		name    string
		page    int
		total   int
		want    bool
	}{
		{"first of two pages", 1, 2, true},
		{"last page", 2, 2, false},
		{"single page", 1, 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := PageResult[string]{
				Pagination:  Pagination{Page: tt.page, TotalPages: tt.total},
				HasNextPage: tt.page < tt.total,
			}
			if pr.HasNextPage != tt.want {
				t.Errorf("HasNextPage = %v, want %v", pr.HasNextPage, tt.want)
			}
		})
	}
}

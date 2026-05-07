package domain

import "testing"

func TestNewPagination_TotalPagesMinimumOneWhenDataExists(t *testing.T) {
	p := NewPagination(1, 10, 1)
	if p.TotalPages != 1 {
		t.Fatalf("expected 1 total page, got %d", p.TotalPages)
	}
}

func TestNewPagination_TotalPagesZeroWhenNoData(t *testing.T) {
	p := NewPagination(1, 10, 0)
	if p.TotalPages != 0 {
		t.Fatalf("expected 0 total page, got %d", p.TotalPages)
	}
}

package testutil_test

import (
	"testing"

	"github.com/chrishrb/blog-microservice/internal/testutil"
)

func TestPtr(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		s := "test"
		ptr := testutil.Ptr(s)

		if ptr == nil {
			t.Fatal("expected pointer to not be nil")
		}

		if *ptr != s {
			t.Errorf("expected %s, got %s", s, *ptr)
		}
	})

	t.Run("int", func(t *testing.T) {
		i := 42
		ptr := testutil.Ptr(i)

		if ptr == nil {
			t.Fatal("expected pointer to not be nil")
		}

		if *ptr != i {
			t.Errorf("expected %d, got %d", i, *ptr)
		}
	})

	t.Run("struct", func(t *testing.T) {
		type testStruct struct {
			Field string
		}

		s := testStruct{Field: "value"}
		ptr := testutil.Ptr(s)

		if ptr == nil {
			t.Fatal("expected pointer to not be nil")
		}

		if ptr.Field != s.Field {
			t.Errorf("expected %s, got %s", s.Field, ptr.Field)
		}
	})

	t.Run("slice", func(t *testing.T) {
		slice := []int{1, 2, 3}
		ptr := testutil.Ptr(slice)

		if ptr == nil {
			t.Fatal("expected pointer to not be nil")
		}

		if len(*ptr) != len(slice) {
			t.Fatalf("expected length %d, got %d", len(slice), len(*ptr))
		}

		for i, v := range slice {
			if (*ptr)[i] != v {
				t.Errorf("at index %d: expected %d, got %d", i, v, (*ptr)[i])
			}
		}
	})

	t.Run("map", func(t *testing.T) {
		m := map[string]int{"key": 123}
		ptr := testutil.Ptr(m)

		if ptr == nil {
			t.Fatal("expected pointer to not be nil")
		}

		if len(*ptr) != len(m) {
			t.Fatalf("expected length %d, got %d", len(m), len(*ptr))
		}

		if (*ptr)["key"] != m["key"] {
			t.Errorf("expected value %d for key 'key', got %d", m["key"], (*ptr)["key"])
		}
	})
}

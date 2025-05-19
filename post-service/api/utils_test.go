package api

import (
	"testing"

	"github.com/chrishrb/blog-microservice/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestGetPaginationDefaults(t *testing.T) {
	tests := []struct {
		name        string
		paramOffset *int
		paramLimit  *int
		wantOffset  int
		wantLimit   int
	}{
		{
			name:        "nil parameters",
			paramOffset: nil,
			paramLimit:  nil,
			wantOffset:  0,
			wantLimit:   20,
		},
		{
			name:        "custom offset",
			paramOffset: testutil.Ptr(5),
			paramLimit:  nil,
			wantOffset:  5,
			wantLimit:   20,
		},
		{
			name:        "custom limit",
			paramOffset: nil,
			paramLimit:  testutil.Ptr(30),
			wantOffset:  0,
			wantLimit:   30,
		},
		{
			name:        "custom offset and limit",
			paramOffset: testutil.Ptr(10),
			paramLimit:  testutil.Ptr(50),
			wantOffset:  10,
			wantLimit:   50,
		},
		{
			name:        "limit exceeding maximum",
			paramOffset: testutil.Ptr(0),
			paramLimit:  testutil.Ptr(200),
			wantOffset:  0,
			wantLimit:   100,
		},
		{
			name:        "limit at maximum",
			paramOffset: nil,
			paramLimit:  testutil.Ptr(100),
			wantOffset:  0,
			wantLimit:   100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOffset, gotLimit := getPaginationWithDefaults(tt.paramOffset, tt.paramLimit)
			assert.Equal(t, tt.wantOffset, gotOffset)
			assert.Equal(t, tt.wantLimit, gotLimit)
		})
	}
}

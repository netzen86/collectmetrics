package handlers

import (
	"reflect"
	"testing"

	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
)

func TestGetMetrics(t *testing.T) {
	// type url struct {
	// 	url string
	// }
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name string
		url  string
		want want
	}{
		// {
		// 	name: "test gauge",
		// 	url:  "update/gauge/Alloc/100",
		// 	want: want{
		// 		code:        http.StatusOK,
		// 		response:    `{"status":"ok"}`,
		// 		contentType: "text/html",
		// 	},
		// },
	}
	storage, err := memstorage.NewMemStorage()
	if err != nil {
		t.Errorf("expected err to be nil got %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// httptest.NewRequest(http.MethodPost, tt.url, nil)

			if got := GetMetrics(storage); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

package handlers

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/netzen86/collectmetrics/internal/repositories/memstorage"
)

func TestGetMetrics(t *testing.T) {
	type args struct {
		storage *memstorage.MemStorage
	}
	tests := []struct {
		name string
		args args
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMetrics(tt.args.storage); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

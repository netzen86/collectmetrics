package handlers

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/netzen86/collectmetrics/internal/repositories"
)

func TestUpdateMHandle(t *testing.T) {
	type args struct {
		storage repositories.Repo
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
			if got := UpdateMHandle(tt.args.storage, "test", "test", "test"); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateMHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

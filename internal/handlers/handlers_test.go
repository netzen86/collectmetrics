package handlers

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/netzen86/collectmetrics/internal/repositories"
	"go.uber.org/zap"
)

func TestUpdateMHandle(t *testing.T) {
	type args struct {
		storage repositories.Repo
		logger  zap.SugaredLogger
	}
	tests := []struct {
		args args
		want http.HandlerFunc
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UpdateMHandle(tt.args.storage, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpdateMHandle() = %v, want %v", got, tt.want)
			}
		})
	}
}

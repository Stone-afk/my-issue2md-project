package fetchprovider

import (
	"context"
	"errors"
	"testing"

	"github.com/stoneafk/issue2md/internal/model"
	"github.com/stoneafk/issue2md/internal/parser"
)

type stubProvider struct{}

func (stubProvider) Fetch(context.Context, parser.Target, model.FetchOptions) (model.DocumentData, error) {
	return model.DocumentData{}, nil
}

func TestProviderContractAndErrors(t *testing.T) {
	t.Run("provider exposes fetch only", func(t *testing.T) {
		var provider Provider = stubProvider{}
		if _, ok := any(provider).(interface {
			GetIssue(context.Context, parser.Target, model.FetchOptions) (model.DocumentData, error)
		}); ok {
			t.Fatal("provider interface should not require GetIssue")
		}
	})

	tests := []struct {
		name                        string
		err                         error
		wantIsProviderNotRegistered bool
		wantIsUnsupported           bool
		wantAsUnsupported           bool
		wantUnsupportedProvider     model.Provider
		wantUnsupportedKind         model.ContentKind
	}{
		{
			name:                        "provider not registered is distinguishable",
			err:                         ErrProviderNotRegistered,
			wantIsProviderNotRegistered: true,
			wantIsUnsupported:           false,
			wantAsUnsupported:           false,
		},
		{
			name:                        "unsupported capability is distinguishable",
			err:                         UnsupportedCapability(model.ProviderGitLab, model.KindDiscussion),
			wantIsProviderNotRegistered: false,
			wantIsUnsupported:           true,
			wantAsUnsupported:           true,
			wantUnsupportedProvider:     model.ProviderGitLab,
			wantUnsupportedKind:         model.KindDiscussion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Is(tt.err, ErrProviderNotRegistered); got != tt.wantIsProviderNotRegistered {
				t.Fatalf("errors.Is(err, ErrProviderNotRegistered) = %v, want %v", got, tt.wantIsProviderNotRegistered)
			}
			if got := IsUnsupportedCapability(tt.err); got != tt.wantIsUnsupported {
				t.Fatalf("IsUnsupportedCapability(err) = %v, want %v", got, tt.wantIsUnsupported)
			}

			var unsupported *UnsupportedCapabilityError
			gotAsUnsupported := errors.As(tt.err, &unsupported)
			if gotAsUnsupported != tt.wantAsUnsupported {
				t.Fatalf("errors.As(err, *UnsupportedCapabilityError) = %v, want %v", gotAsUnsupported, tt.wantAsUnsupported)
			}
			if tt.wantAsUnsupported {
				if unsupported.Provider != tt.wantUnsupportedProvider {
					t.Fatalf("Provider = %q, want %q", unsupported.Provider, tt.wantUnsupportedProvider)
				}
				if unsupported.Kind != tt.wantUnsupportedKind {
					t.Fatalf("Kind = %q, want %q", unsupported.Kind, tt.wantUnsupportedKind)
				}
			}
		})
	}
}

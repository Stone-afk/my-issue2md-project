package fetchprovider

import (
	"context"
	"errors"
	"fmt"

	"github.com/stoneafk/issue2md/internal/model"
	"github.com/stoneafk/issue2md/internal/parser"
)

type Provider interface {
	Fetch(ctx context.Context, target parser.Target, opts model.FetchOptions) (model.DocumentData, error)
}

var ErrProviderNotRegistered = errors.New("provider not registered")

type UnsupportedCapabilityError struct {
	Provider model.Provider
	Kind     model.ContentKind
}

func (e *UnsupportedCapabilityError) Error() string {
	return fmt.Sprintf("provider %q does not support kind %q", e.Provider, e.Kind)
}

func UnsupportedCapability(provider model.Provider, kind model.ContentKind) error {
	return &UnsupportedCapabilityError{Provider: provider, Kind: kind}
}

func IsUnsupportedCapability(err error) bool {
	var unsupported *UnsupportedCapabilityError
	return errors.As(err, &unsupported)
}

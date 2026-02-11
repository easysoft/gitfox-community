package logs

import (
	"context"
	"io"
)

// Interface provides an interface for the persistent log store backend.
type Interface interface {
	// Find returns a log stream from the datastore.
	Find(ctx context.Context, filePath string) (io.ReadCloser, error)

	// Create writes copies the log stream from Reader r to the datastore.
	Create(ctx context.Context, filePath string, r io.Reader) error

	// Update copies the log stream from Reader r to the datastore.
	Update(ctx context.Context, filePath string, r io.Reader) error

	// Delete purges the log stream from the datastore.
	Delete(ctx context.Context, filePath string) error
}

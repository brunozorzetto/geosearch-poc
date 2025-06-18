package vertexai

import (
	"context"

	retail "cloud.google.com/go/retail/apiv2/retailpb"
)

// VertexAIClientInterface defines the interface for Vertex AI operations
type VertexAIClientInterface interface {
	Search(ctx context.Context, params SearchParams) ([]SearchResult, string, error)
	CreateProduct(ctx context.Context, product *retail.Product) error
	UpdateProduct(ctx context.Context, product *retail.Product) error
	DeleteProduct(ctx context.Context, productID string) error
	Close() error
}

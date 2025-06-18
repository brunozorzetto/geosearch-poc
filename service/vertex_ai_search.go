package service

import (
	"context"
	"fmt"

	retail "cloud.google.com/go/retail/apiv2"
	"cloud.google.com/go/retail/apiv2/retailpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type VertexAISearch struct {
	client    *retail.SearchClient
	projectID string
	location  string
}

func NewVertexAISearch(projectID, location string, credentialsFile string) (*VertexAISearch, error) {
	ctx := context.Background()
	client, err := retail.NewSearchClient(ctx, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("failed to create search client: %w", err)
	}

	return &VertexAISearch{
		client:    client,
		projectID: projectID,
		location:  location,
	}, nil
}

func (s *VertexAISearch) Search(ctx context.Context, query, filter string, pageSize int, pageToken string) ([]*retailpb.SearchResponse_SearchResult, string, error) {
	req := &retailpb.SearchRequest{
		Placement: fmt.Sprintf("projects/%s/locations/%s/catalogs/default_catalog/placements/default_search", s.projectID, s.location),
		Query:     query,
		Filter:    filter,
		PageSize:  int32(pageSize),
		PageToken: pageToken,
	}

	iter := s.client.Search(ctx, req)
	results := make([]*retailpb.SearchResponse_SearchResult, 0)

	for {
		result, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, "", fmt.Errorf("failed to get next result: %w", err)
		}
		results = append(results, result)
	}

	return results, iter.PageInfo().Token, nil
}

func (s *VertexAISearch) Close() error {
	return s.client.Close()
}

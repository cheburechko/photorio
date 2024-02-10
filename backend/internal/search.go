package internal

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
)

type (
	SearchItem struct {
		URL     string `json:"url"`
		Caption string `json:"caption"`
	}

	SearchResponse struct {
		Items []*SearchItem
	}
)

func Search(query string, es *elasticsearch.TypedClient, context context.Context) (*SearchResponse, error) {
	result, err := es.Search().
		Index("captions").
		Request(&search.Request{
			Query: &types.Query{
				Match: map[string]types.MatchQuery{
					"caption": {Query: query},
				},
			},
		}).
		Do(context)

	if err != nil {
		return nil, err
	}

	resp := SearchResponse{
		Items: make([]*SearchItem, result.Hits.Total.Value),
	}

	for i, item := range result.Hits.Hits {
		var parsed SearchItem
		err = json.Unmarshal(item.Source_, &parsed)
		if err != nil {
			marshalled, _ := item.Source_.MarshalJSON()
			slog.Error("failed to unmarshall json", slog.String("json", string(marshalled)))
			return nil, err
		}
		resp.Items[i] = &parsed
	}

	return &resp, nil
}

package utils

import (
	"context"
	"encoding/json"
	log "github.com/cihub/seelog"
	"github.com/olivere/elastic/v7"
	"wilikidi/es/model"
)

type Entry interface {
	SearchAggregation() ([]string, error)
	SearchAggregationByFlag(knowFlags []string) ([]string, error)

	SearchEntryByFlag(flags []string) ([]model.EntryItem, error)
	DeleteEntryById(id string) error

	AppendEntry(item model.EntryItem) (string, error)
	UpdateEntry(id string, item model.EntryItem) error
}

type SearchEntry struct {
}

var (
	searchEntryIndex = "entry"
	se               SearchEntry
)

// SearchAggregationByFlag 按照已知 flag 进行聚合。
func (s *SearchEntry) SearchAggregationByFlag(knowFlags []string) ([]string, error) {
	var flags []string
	knowFlagsMap := make(map[string]struct{})

	query := elastic.NewBoolQuery()
	for _, flag := range knowFlags {
		knowFlagsMap[flag] = struct{}{}
		flagQ := elastic.NewTermQuery("flag", flag)
		query.Should(flagQ)
	}

	flagAgg := elastic.NewTermsAggregation().Field("flag.keyword")

	aggResp, err := ESCLI.Search().
		Index(searchEntryIndex).
		Query(query).
		Aggregation("uniq_flags", flagAgg).
		Size(0).
		Do(context.Background())

	if err != nil {
		log.Errorf("[SearchAggregation] search aggregation failed: %v", err)
		return nil, err
	}

	aggInfo, _ := aggResp.Aggregations.Terms("uniq_flags")
	aggBucket := aggInfo.Aggregations["buckets"]

	var flagsBucket []FlagAgg
	err = json.Unmarshal(aggBucket, &flagsBucket)

	if err != nil {
		log.Errorf("[SearchAggregation] search aggregation failed: %v", err)
		return nil, err
	}

	for _, flag := range flagsBucket {
		// 过滤查询的 flag
		if _, ok := knowFlagsMap[flag.Key]; ok {
			continue
		}
		flags = append(flags, flag.Key)
	}

	return flags, nil
}

// SearchAggregation 搜索所有的 flag 聚合结果
func (s *SearchEntry) SearchAggregation() ([]string, error) {
	var flags []string

	flagAgg := elastic.NewTermsAggregation().Field("flag.keyword")

	aggResp, err := ESCLI.Search().
		Index(searchEntryIndex).
		Aggregation("uniq_flags", flagAgg).
		Size(0).
		Do(context.Background())

	if err != nil {
		log.Errorf("[SearchAggregation] search aggregation failed: %v", err)
		return nil, err
	}

	aggInfo, _ := aggResp.Aggregations.Terms("uniq_flags")
	aggBucket := aggInfo.Aggregations["buckets"]

	var flagsBucket []FlagAgg
	err = json.Unmarshal(aggBucket, &flagsBucket)

	if err != nil {
		log.Errorf("[SearchAggregation] search aggregation failed: %v", err)
		return nil, err
	}

	for _, flag := range flagsBucket {
		flags = append(flags, flag.Key)
	}

	return flags, nil
}

func (s *SearchEntry) SearchEntryByFlag(flags []string) (items []model.EntryItem, err error) {
	query := elastic.NewBoolQuery()

	// 使用多个 should 增加评分排序
	for _, flag := range flags {
		flagQ := elastic.NewTermQuery("flag", flag)
		query.Should(flagQ)
	}

	resp, err := ESCLI.Search().
		Index(searchEntryIndex).
		Query(query).
		Do(context.Background())

	if err != nil {
		log.Errorf("[SearchEntryByFlag] search entry failed: %v", err)
		return nil, err
	}

	if resp.Hits.TotalHits.Value == 0 || len(resp.Hits.Hits) == 0 {
		log.Warnf("[SearchEntryByFlag] search entry failed: %v", err)
		return nil, nil
	}

	for _, hit := range resp.Hits.Hits {
		var item model.EntryItem
		// don't care about the error
		json.Unmarshal(hit.Source, &item)

		items = append(items, item)
	}

	return
}

func (s *SearchEntry) DeleteEntryById(id string) error {
	_, err := ESCLI.Delete().Index(searchEntryIndex).Id(id).Do(context.Background())
	if err != nil {
		log.Errorf("[DeleteEntryById] delete entry failed: %v", err)
		return err
	}
	return nil
}

func (s *SearchEntry) AppendEntry(item model.EntryItem) (string, error) {
	resp, err := ESCLI.Index().
		Index(searchEntryIndex).
		BodyJson(item).
		Refresh("true").
		Do(context.Background())

	if err != nil {
		log.Errorf("[AppendEntry] append entry failed: %v", err)
		return "", err
	}

	return resp.Id, nil
}

func (s *SearchEntry) UpdateEntry(id string, item model.EntryItem) error {
	_, err := ESCLI.Index().
		Index(searchEntryIndex).
		Id(id).
		BodyJson(item).
		Refresh("true").
		Do(context.Background())

	if err != nil {
		log.Errorf("[UpdateEntry] update entry failed: %v", err)
		return err
	}

	return nil
}

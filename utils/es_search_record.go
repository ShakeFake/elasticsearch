package utils

import (
	"context"
	"encoding/json"
	log "github.com/cihub/seelog"
	"github.com/olivere/elastic/v7"
	"wilikidi/es/model"
)

type Record interface {
	AppendRecord(info interface{}) (err error, id string)
	UpdateRecord(id string, info interface{}) error
	UpdateRecordFlags(id string, flags []string) error
	UpdateRecordCommand(id string, command string) error

	SearchRecordByFlag(flags []string) []model.SearchRecordItem
	SearchRecordById(id string) (item model.SearchRecordItem, err error)

	DeleteRecordById(id string) error

	AggregationFlags() []string
}

type SearchRecord struct {
}

var (
	searchRecordIndex = "record"
	sr                SearchRecord
)

func AppendMultiData(items []model.SearchRecordItem) {
	for _, item := range items {
		err, _ := sr.AppendRecord(item)
		if err != nil {
			log.Errorf("[AppendMultiData] append record %v failed: %v", item, err)
			continue
		}
	}
}

// AppendRecord 追加数据
func (s *SearchRecord) AppendRecord(info interface{}) (err error, id string) {
	resp, err := ESCLI.Index().
		Index(searchRecordIndex).
		BodyJson(info).
		Refresh("true").
		Do(context.Background())
	if err != nil {
		log.Errorf("[AppendRecord] append record failed: %v", err)
		return err, ""
	}

	return nil, resp.Id
}

// UpdateRecord 注意这个更新，以传入的 info 为准。
func (s *SearchRecord) UpdateRecord(id string, info interface{}) error {
	_, err := ESCLI.Index().
		Index(searchRecordIndex).
		Id(id).
		BodyJson(info).
		Refresh("true").
		Do(context.Background())
	if err != nil {
		log.Errorf("[UpdateRecord] update record failed: %v", err)
		return err
	}

	return nil
}

// UpdateRecordFlags 更新 flag 信息。
func (s *SearchRecord) UpdateRecordFlags(id string, flags []string) error {
	item, err := s.SearchRecordById(id)
	if err != nil {
		return err
	}

	item.Flag = flags

	err = s.UpdateRecord(id, item)
	if err != nil {
		return err
	}
	return nil
}

// UpdateRecordCommand 更新 command 信息。
func (s *SearchRecord) UpdateRecordCommand(id string, command string) error {
	item, err := s.SearchRecordById(id)
	if err != nil {
		return err
	}

	item.Command = command

	err = s.UpdateRecord(id, item)

	if err != nil {
		return err
	}
	return nil
}

// SearchRecordByFlag 根据 flag 搜索数据
func (s *SearchRecord) SearchRecordByFlag(flags []string) (items []model.SearchRecordItem) {
	query := elastic.NewBoolQuery()

	// 使用多个 should 按照默认评分进行排序。
	for _, flag := range flags {
		flagQ := elastic.NewTermQuery("flag", flag)
		query.Should(flagQ)
	}

	res, err := ESCLI.Search().
		Index(searchRecordIndex).
		Query(query).
		Do(context.Background())

	if err != nil {
		log.Errorf("[SearchRecordByFlag] search record failed: %v", err)
		return nil
	}
	if res.Hits.TotalHits.Value <= 0 || len(res.Hits.Hits) <= 0 {
		log.Warnf("[SearchRecordByFlag] search record is zero")
		return nil
	}

	for _, hit := range res.Hits.Hits {
		var item model.SearchRecordItem
		json.Unmarshal(hit.Source, &item)

		items = append(items, item)
	}

	return items
}

// todo: 引入分词器，引入 embedding 如何？

// SearchRecordById 根据 id 搜索数据。
func (s *SearchRecord) SearchRecordById(id string) (item model.SearchRecordItem, err error) {
	query := elastic.NewBoolQuery()
	idQ := elastic.NewTermQuery("_id", id)
	query.Must(idQ)

	resp, err := ESCLI.Search().Index(searchRecordIndex).Query(query).Do(context.Background())
	if err != nil {
		log.Errorf("[SearchRecordById] search record failed: %v", err)
		return item, err
	}

	if resp.Hits.TotalHits.Value <= 0 || len(resp.Hits.Hits) <= 0 {
		log.Warnf("[SearchRecordById] search record is zero")
		return item, err
	}

	for _, hit := range resp.Hits.Hits {
		json.Unmarshal(hit.Source, &item)
	}

	return item, nil
}

func (s *SearchRecord) DeleteRecordById(id string) error {
	_, err := ESCLI.Delete().
		Index(searchRecordIndex).
		Id(id).
		Do(context.Background())
	if err != nil {
		log.Errorf("[DeleteRecordById] delete record failed: %v", err)
		return err
	}

	return nil
}

type FlagAgg struct {
	Key      string `json:"key"`
	DocCount int    `json:"doc_count"`
}

func (s *SearchRecord) AggregationFlags() []string {
	var flags []string

	flagAgg := elastic.NewTermsAggregation().Field("flag.keyword")

	aggResp, err := ESCLI.Search().
		Index(searchRecordIndex).
		Aggregation("uniq_flags", flagAgg).
		Size(0).
		Do(context.Background())

	if err != nil {
		log.Infof("[AggregationFlags] aggregation flags failed: %v", err)
		return nil
	}

	aggInfo, _ := aggResp.Aggregations.Terms("uniq_flags")
	aggBucket := aggInfo.Aggregations["buckets"]

	var flagsBucket []FlagAgg
	err = json.Unmarshal(aggBucket, &flagsBucket)
	if err != nil {
		log.Infof("[AggregationFlags] unmarshal aggregation flags failed: %v", err)
		return nil
	}

	for _, flag := range flagsBucket {
		flags = append(flags, flag.Key)
	}

	return flags
}

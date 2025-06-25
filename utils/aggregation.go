package utils

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/olivere/elastic/v7"
	"wilikidi/es/model"
)

func GetAggregationByName(index string) (error, map[string]string) {
	agg := elastic.NewTermsAggregation().Field("name.keyword").Size(ESMAXSIZE)

	err, aggBucket := Aggregation(index, agg)
	if err != nil {
		return err, nil
	}

	// 针对不同的构造，最后都是构造 bucket 的结构。
	err, result := onceAggregationUnmarshal(aggBucket)
	if err != nil {
		log.Infof("[GetAggregation] get aggregation failed: %v", err)
		return err, nil
	}

	aggMap, _ := constructAggToMapAndKeys(result)
	return nil, aggMap

}

type AggregationSum struct {
	Key      string `json:"key"`
	DocCount int    `json:"doc_count"`
	AgeSum   AgeSum `json:"age_sum"`
}

type AgeSum struct {
	Value float64 `json:"value"`
}

func GetAggregationByNameSumAge(index string) (error, map[string]string) {
	ageSumAgg := elastic.NewSumAggregation().Field("age")
	agg := elastic.NewTermsAggregation().Field("name.keyword").Size(ESMAXSIZE).SubAggregation("age_sum", ageSumAgg)

	err, aggBucket := Aggregation(index, agg)
	if err != nil {
		log.Infof("[GetAggregation] get aggregation failed: %v", err)
		return err, nil
	}
	var aggSum []AggregationSum
	err = json.Unmarshal(aggBucket, &aggSum)
	if err != nil {
		return err, nil
	}

	aggMap := make(map[string]string, 0)
	for _, item := range aggSum {
		n, a := BigParse(fmt.Sprintf("%v", item.AgeSum.Value))
		_ = a
		aggMap[item.Key] = n
	}

	return nil, aggMap
}

type NameAgeAgg struct {
	Key       string    `json:"key"`
	Value     int       `json:"value"`
	UniqueAge AgeBucket `json:"unique_age"`
}

type AgeBucket struct {
	Buckets []AgeAgg `json:"buckets"`
}

type AgeAgg struct {
	Key      int `json:"key"`
	DocCount int `json:"doc_count"`
}

type ReturnData struct {
	Name    string   `json:"name"`
	AgeInfo []string `json:"age_info"`
}

func GetDoubleA(index string) (error, []ReturnData) {
	ageAggregation := elastic.NewTermsAggregation().
		Field("age").
		Size(ESMAXSIZE)
	nameAggregation := elastic.NewTermsAggregation().
		Field("name.keyword").
		Size(ESMAXSIZE).
		SubAggregation("unique_age", ageAggregation)

	err, resp := Aggregation(index, nameAggregation)

	if err != nil {
		return err, nil
	}

	var nameAgeInfo []NameAgeAgg
	err = json.Unmarshal(resp, &nameAgeInfo)
	if err != nil {
		return err, nil
	}

	returnDates := make([]ReturnData, 0)
	for _, nameItem := range nameAgeInfo {
		returnData := ReturnData{}
		returnData.Name = nameItem.Key
		for _, ageItem := range nameItem.UniqueAge.Buckets {
			returnData.AgeInfo = append(
				returnData.AgeInfo,
				fmt.Sprintf("age is: %v, number is: %v", ageItem.Key, ageItem.DocCount),
			)
		}

		returnDates = append(returnDates, returnData)
	}

	return nil, returnDates

}

func Aggregation(index string, agg *elastic.TermsAggregation) (error, []byte) {

	aggResp, err := ESCLI.Search().
		Index(index).
		Aggregation("unique_agg", agg).
		Size(0).
		Do(context.Background())

	if err != nil {
		log.Infof("[GetAggregation] search es failed: %v", err)
		return err, nil
	}

	aggInfo, _ := aggResp.Aggregations.Terms("unique_agg")
	aggBucket := aggInfo.Aggregations["buckets"]

	return nil, aggBucket
}

func onceAggregationUnmarshal(aggInfo []byte) (error, []model.OnceAggregation) {
	var agg = make([]model.OnceAggregation, 0)

	err := json.Unmarshal(aggInfo, &agg)
	if err != nil {
		return err, nil
	} else {
		return nil, agg
	}
}

func constructAggToMapAndKeys(items []model.OnceAggregation) (map[string]string, []string) {
	var aggMap = make(map[string]string, 0)
	var allTask = make([]string, 0)
	for _, item := range items {
		aggMap[item.Key] = ""
		allTask = append(allTask, item.Key)
	}
	return aggMap, allTask
}

package utils

import (
	"context"
	"crypto/tls"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/olivere/elastic/v7"
	"net/http"
	"strings"
	"wilikidi/es/model"
)

var (
	ESCLI *elastic.Client
)

// InitES Init ES Connect
func InitES() {
	client, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(strings.Split(ES_HOST, ";")...),
		elastic.SetBasicAuth(ESUSER, ESPASSWD),
		elastic.SetHttpClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}),
	)
	if err != nil {
		panic(err)
	}

	ESCLI = client
}

func InsertByIndex(index string, data interface{}) error {
	_, err := ESCLI.Index().
		Index(index).
		BodyJson(data).
		Do(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func UpdateById(index string, ids []string, field string, value interface{}) error {
	return Update(index, "_id", ids, field, value)
}

func Update(index string, key string, ids []string, field string, value interface{}) error {

	query := elastic.NewBoolQuery()
	query.Must(elastic.NewTermsQuery(key, UnWrapper(ids)...))

	dsl, _ := query.Source()
	ShowDSL(dsl)

	script := elastic.NewScript(
		fmt.Sprintf("ctx._source['%v']", field)+"=params['value_insert']",
	).Param("value_insert", value)

	_, err := ESCLI.UpdateByQuery().
		Index(index).
		Query(query).
		Script(script).
		Refresh("true").
		Do(context.Background())

	if err != nil {
		log.Infof("[UpdateById] update data failed: %v", err)
	}

	return nil
}

func DeleteById(index string, ids []string) error {
	return Delete(index, "_id", ids)
}

func Delete(index string, key string, ids []string) error {

	query := elastic.NewBoolQuery()
	query.Must(elastic.NewTermsQuery(key, UnWrapper(ids)...))

	_, err := ESCLI.DeleteByQuery(index).Query(query).Do(context.Background())
	if err != nil {
		log.Infof("[DeleteById] delete data failed: %v", err)
		return err
	}

	return nil

}

func SearchById(index string, ids []string) (error, interface{}) {
	return Search(index, "_id", ids)
}

func Search(index string, key string, ids []string) (error, interface{}) {

	query := elastic.NewBoolQuery()
	query.Must(elastic.NewTermsQuery(key, UnWrapper(ids)...))

	resp, err := ESCLI.Search().
		Index(index).
		Query(query).
		Do(context.Background())

	if err != nil {
		log.Infof("[SearchById] search es failed: %v", err)
		return err, nil
	}

	var students []model.Student
	for _, hit := range resp.Hits.Hits {
		var student model.Student
		GetUnmarshal(string(hit.Source), &student)
		students = append(students, student)
	}

	return nil, students

}

// Collapse collapse 和 agg 不同。agg 不改变结果size。前者改变。
func Collapse(index string) []ReturnData {
	ageCollapse := elastic.NewCollapseBuilder("age")

	nameCollapse := elastic.
		NewCollapseBuilder("name.keyword").
		InnerHit(
			elastic.NewInnerHit().
				Name("age_collapse").
				Size(ESMAXSIZE).
				Collapse(ageCollapse).
				Sort("age", false),
		)

	resp, err := ESCLI.Search().
		Index(index).
		From(1).
		Size(10).
		Collapse(nameCollapse).Do(context.Background())

	if err != nil {
		log.Infof("[Collapse] search es failed: %v", err)
		return nil
	}

	var returnDatas []ReturnData
	for _, hit := range resp.Hits.Hits {
		var returnData ReturnData
		var student model.Student

		err := GetUnmarshal(string(hit.Source), &student)
		if err != nil {
			log.Infof("unmarshal failed: %v", err)
			continue
		}
		returnData.Name = student.Name

		innerHits := hit.InnerHits["age_collapse"].Hits.Hits
		if len(innerHits) == 0 {
			log.Warnf("%v innerHits is empty", student.Name)
			continue
		}

		for _, innerHit := range innerHits {

			err = GetUnmarshal(string(innerHit.Source), &student)
			if err != nil {
				log.Infof("unmarshal failed: %v", err)
				continue
			}

			returnData.AgeInfo = append(returnData.AgeInfo, fmt.Sprintf("pre one, age: %v \n", student.Age))
		}

		returnDatas = append(returnDatas, returnData)
	}

	return returnDatas
}

func QueryDsl() {
	pathQuery := elastic.NewBoolQuery().Must(elastic.NewTermsQuery("placeholders.path", "abc"))
	qStory := elastic.NewNestedQuery("placeholders", pathQuery).InnerHit(elastic.NewInnerHit().From(0).Size(100))

	dsl, _ := qStory.Source()
	fmt.Println(ShowDSL(dsl))

}

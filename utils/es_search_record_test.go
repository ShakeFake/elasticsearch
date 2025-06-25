package utils

import (
	"fmt"
	"testing"
	"wilikidi/es/Docker/es"
	"wilikidi/es/model"
)

func Init() {
	ES_HOST = "https://192.168.1.1:9200"
	ESUSER = "elastic"
	ESPASSWD = "jE01pNvu9v-XOubIs5na"
}

func TestAppendMultiData(t *testing.T) {
	Init()
	InitES()

	AppendMultiData(es.CommandDev)

}

func TestSearchRecord_AppendRecord(t *testing.T) {
	Init()
	InitES()

	var item model.SearchRecordItem
	item.Flag = []string{"cluter", "health"}
	item.Command = "GET  _cat/health"

	err, id := sr.AppendRecord(item)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)
}

func TestSearchRecord_UpdateRecord(t *testing.T) {
	Init()
	InitES()

	var item model.SearchRecordItem
	item.Flag = []string{"mma_update", "special_path_update"}

	id := "e399hI4BbUF59OgB8sti"
	err := sr.UpdateRecord(id, item)
	if err != nil {
		panic(err)
	}
}

func TestSearchRecord_SearchRecordByFlag(t *testing.T) {
	Init()
	InitES()

	flags := []string{"mma", "special_path"}

	result := sr.SearchRecordByFlag(flags)

	fmt.Println(result)
}

func TestSearchRecord_SearchRecordById(t *testing.T) {
	Init()
	InitES()

	id := "e399hI4BbUF59OgB8sti"

	result, err := sr.SearchRecordById(id)
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}

func TestSearchRecord_DeleteRecordById(t *testing.T) {
	Init()
	InitES()

	id := "e399hI4BbUF59OgB8sti"

	err := sr.DeleteRecordById(id)
	if err != nil {
		panic(err)
	}
}

func TestSearchRecord_AggregationFlags(t *testing.T) {
	Init()
	InitES()

	result := sr.AggregationFlags()
	fmt.Println(result)
}

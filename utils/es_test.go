package utils

import (
	"errors"
	"fmt"
	log "github.com/cihub/seelog"
	"testing"
	"wilikidi/es/model"
)

func Local() {
	ES_HOST = "https://192.168.1.1:9200"
	ESMAXSIZE = 10000
	ESUSER = "elastic"
	ESPASSWD = "bWGOpD5yK0En0-HvQ+8Y"
}

func init() {
	Local()
}

func getTestSample() model.Student {
	student := model.Student{
		Name: "abc",
		Age:  18,
	}

	return student
}

func TestHundred(t *testing.T) {
	InitES()

	student := getTestSample()
	for i := 0; i < 100; i++ {
		student.Name = "abc" + fmt.Sprintf("%v", i/10)
		student.Age = i
		err := InsertByIndex("yunkai", student)
		if err != nil {
			panic(err)

		}
	}
}

func TestInsertByIndex(t *testing.T) {
	InitES()

	student := getTestSample()
	err := InsertByIndex("yunkai", student)
	if err != nil {
		panic(err)
	}
}

func TestUpdateScript(t *testing.T) {
	InitES()

	ids := []string{"HXcIS4sBlB1W240i8zJm"}

	// 更新不存在的field，会创建新的field。如果类型不同，也会更新，会把对应的 field 更新掉。
	err := UpdateById("yunkai", ids, "name", "abc")
	if err != nil {
		panic(err)
	}
}

func TestDeleteById(t *testing.T) {
	InitES()

	ids := []string{"HXcIS4sBlB1W240i8zJm"}
	err := DeleteById("yunkai", ids)

	if err != nil {
		panic(err)
	}
}

func TestSearchByIndex(t *testing.T) {
	InitES()

	ids := []string{"HXcIS4sBlB1W240i8zJm"}
	err, result := SearchById("yunkai", ids)
	if err != nil {
		panic(err)
	}

	log.Infof("result: %v", result)
}

func TestCollapse(t *testing.T) {
	InitES()

	result := Collapse("yunkai")

	if result == nil {
		panic(errors.New("result is nil"))
	}

	for k, v := range result {
		fmt.Printf("%v-%v\n", k, v)
	}
}

func TestQueryDsl(t *testing.T) {
	QueryDsl()
}

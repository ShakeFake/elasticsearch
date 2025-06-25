package utils

import (
	"fmt"
	"testing"
	"wilikidi/es/model"
)

func TestAppendData(t *testing.T) {
	Init()
	InitES()

	item := model.EntryItem{
		Flag: []string{"abc", "ab", "abcd"},
		Link: "qwe",
	}

	id, err := se.AppendEntry(item)
	if err != nil {
		panic(err)
	}

	fmt.Println(id)
}

func TestSearchEntry_DeleteEntryById(t *testing.T) {
	Init()
	InitES()
	id := "hX-nmI4BbUF59OgBXstm"

	err := se.DeleteEntryById(id)

	if err != nil {
		panic(err)
	}
}

func TestSearchEntry_SearchEntryByFlag(t *testing.T) {
	Init()
	InitES()

	flags := []string{"offcial_doc"}

	result, err := se.SearchEntryByFlag(flags)
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}

func TestSearchEntry_UpdateEntry(t *testing.T) {
	Init()
	InitES()

	flagsSearch := []string{"offcial_doc"}
	result, err := se.SearchEntryByFlag(flagsSearch)
	if err != nil {
		panic(err)
	}

	item := result[0]
	item.Link = fmt.Sprintf("%v_update", item.Link)
	id := "hX-nmI4BbUF59OgBXstm"

	err = se.UpdateEntry(id, item)
	if err != nil {
		panic(err)
	}

	fmt.Println("update success")
}

func TestSearchEntry_SearchAggregationByFlag(t *testing.T) {
	Init()
	InitES()

	knowFlags := []string{"abcd", "tensorflow"}

	resultFlags, err := se.SearchAggregationByFlag(knowFlags)
	if err != nil {
		panic(err)
	}

	fmt.Println(resultFlags)
}

func TestSearchEntry_SearchAggregation(t *testing.T) {
	Init()
	InitES()

	resultFlags, err := se.SearchAggregation()
	if err != nil {
		panic(err)
	}

	fmt.Println(resultFlags)
}

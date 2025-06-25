package utils

import (
	"fmt"
	"testing"
)

var (
	index = "yunkai"
)

func TestGetAggregationByName(t *testing.T) {
	InitES()

	err, result := GetAggregationByName(index)
	if err != nil {
		panic(err)
	}
	for k, v := range result {
		fmt.Printf("%v-%v\n", k, v)
	}

}

func TestGetAggregationByNameSumAge(t *testing.T) {
	InitES()

	err, result := GetAggregationByNameSumAge(index)

	if err != nil {
		panic(err)
	}

	for k, v := range result {
		fmt.Printf("%v-%v\n", k, v)
	}
}

func TestGetDoubleA(t *testing.T) {
	InitES()

	err, result := GetDoubleA(index)

	if err != nil {
		panic(err)
	}

	for k, v := range result {
		fmt.Printf("%v-%v\n", k, v)
	}
}

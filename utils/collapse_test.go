package utils

import (
	"fmt"
	"testing"
	"time"
)

func InitLine() {
	ES_HOST = "https://k8s-default-opensear-18ae5c0dc6-f878aec859fb0636.elb.us-east-1.amazonaws.com"
	ESUSER = "tvu_opensearch_production"
	ESPASSWD = "MMAopensearch#2023"
}

func TestGetPilotSatisTask(t *testing.T) {
	InitLogger()
	InitLine()
	InitES()

	now := time.Now().UnixMilli()
	end := now - oneDay
	start := end - (oneDay * 90)

	GetPilotSatisTask(80, start, end)
	//for key, item := range items {
	//	log.Infof("origin data: key is: %v, value is: %v", key, marshal(item))
	//}
	//
	//results := getAllTask(items)
	//for key, item := range results {
	//	log.Infof("tasks: key is: %v, value is: %v", key, marshal(item))
	//}
	//
	//signal := make(chan int, 5)
	//wg := sync.WaitGroup{}
	//for key, result := range results {
	//	signal <- 1
	//	wg.Add(1)
	//
	//	go func(key string, result []string) {
	//		CheckTaskInRecording(key, result)
	//
	//		wg.Done()
	//		<-signal
	//	}(key, result)
	//}
	//
	//wg.Wait()
}

func TestCheckBitrate(t *testing.T) {
	fmt.Println(checkVBitrate("5000K", "4000K"))
	fmt.Println(checkVBitrate("5000K", "4801K"))
}

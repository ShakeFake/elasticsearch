package utils

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/olivere/elastic/v7"
	"strconv"
	"strings"
	"sync"
)

type UploadMachine struct {
	TimeStamp int64                    `json:"timestamp,omitempty"`
	TagValue  string                   `json:"tagValue,omitempty"`
	Machine   InstanceInfoAndTaskInfos `json:"machine,omitempty"`
}

//func getAllTask(items map[string][]UploadMachine) (tasks map[string][]string) {
//	taskMap := make(map[string]struct{})
//	tasks = make(map[string][]string)
//
//	for key, value := range items {
//		for _, v := range value {
//			for _, task := range v.Machine.TaskList {
//				ins_task := fmt.Sprintf("%v-%v", key, task.TaskId)
//
//				if _, ok := taskMap[ins_task]; ok {
//					continue
//				}
//
//				taskMap[ins_task] = struct{}{}
//
//				if origin, ok := tasks[key]; ok {
//					tasks[key] = append(origin, task.TaskId)
//				} else {
//					tasks[key] = []string{task.TaskId}
//				}
//			}
//		}
//	}
//
//	return
//}

// 必须要全部 process
func getAllTask(item UploadMachine) []string {
	taskIds := []string{}
	length := len(item.Machine.TaskList)
	for _, task := range item.Machine.TaskList {
		if task.TaskStatus == "Processing" {
			taskIds = append(taskIds, task.TaskId)
		}
	}

	if len(taskIds) >= length {
		return taskIds
	}

	return nil
}

type InstanceInfoAndTaskInfos struct {
	InstanceId         string            `json:"instance_id"`
	InstanceIp         string            `json:"instance_ip"`
	InstanceType       string            `json:"instance_type"`
	InstanceStatus     string            `json:"instance_status"`
	InstanceCreateTime int64             `json:"instance_createtime"`
	InstanceStopTime   int64             `json:"instance_stoptime"`
	Region             string            `json:"az"`
	ImageVersion       string            `json:"image_version"`
	BootTime           int64             `json:"boot_time"`
	Metrics            []MetricMapping   `json:"metrics"`
	ServiceName        string            `json:"service_name"`
	ImageType          string            `json:"image_type"`
	TaskList           []TaskInfosResult `json:"task_list"`
}

type MetricMapping struct {
	MetricName  string  `json:"metric_name"`
	MetricType  string  `json:"metric_type"`
	MetricValue float64 `json:"metric_value"`
}

type TaskInfosResult struct {
	Project         string      `json:"project,omitempty"`
	ProjectEnv      string      `json:"project_env,omitempty"`
	TaskId          string      `json:"task_id,omitempty"`
	TaskStatus      string      `json:"task_status,omitempty"`
	TaskDescription Description `json:"task_description,omitempty"`
}

type Description struct {
	RecErrorCode       string `json:"rec_errcode,omitempty"`
	RecPath            string `json:"rec_path,omitempty"`
	TaskStartTimestamp int64  `json:"task_start_timestamp,omitempty"`
	UserAccount        string `json:"user_account,omitempty"`
	UserId             string `json:"user_id,omitempty"`
	RecType            string `json:"rec_type,omitempty"`
}

var (
	am8       = int64(1730462400000)
	oneDay    = int64(24 * 3600 * 1000)
	nineHours = int64(9 * 3600 * 1000)
)

func GetPilotSatisTask(metricValue int, start int64, end int64) {
	query := elastic.NewBoolQuery()

	metricNameQ := elastic.NewTermsQuery("machine.metrics.metric_name", "task")
	metricValueQ := elastic.NewRangeQuery("machine.metrics.metric_value").Gte(metricValue)
	actionQ := elastic.NewTermsQuery("machine.action.keyword", "report all instance and taskLists periods")
	// 需要改成美东时间早上 8:00 ~ 5:00
	query.Must(metricNameQ, metricValueQ, actionQ)

	timeShould := elastic.NewBoolQuery()
	for {
		if am8 > start {
			am8 = am8 - oneDay
		} else {
			break
		}
		pm5 := am8 + nineHours

		if am8 > start && am8 < end {
			timestampGt := elastic.NewRangeQuery("timestamp").Gte(am8).Lte(pm5)
			timeShould.Should(timestampGt)
		}
	}
	query.Must(timeShould)

	//channelQ := elastic.NewTermsQuery("machine.task_list.project.keyword", "Producer")
	//query.Must(channelQ)

	dsl, _ := query.Source()
	log.Infof("[GetPilotSatisTask] dsl is: %v", marshal(dsl))

	//collapseQ := elastic.
	//	NewCollapseBuilder("machine.instance_id.keyword").
	//	InnerHit(
	//		elastic.
	//			NewInnerHit().
	//			Name("sorted").
	//			Size(2000).
	//			Sort("timestamp", true),
	//	)

	// 带上 resolution
	index := "pilot-machine-2024"
	res, err := ESCLI.
		Search().
		Index(index).
		Query(query).
		Size(89999).
		Sort("timestamp", true).
		TrackTotalHits(true).
		Do(context.Background())
	if err != nil {
		log.Infof("[GetPilotSatisTask] search from es failed: %v", err)
		return
	}

	totalData := res.Hits.TotalHits.Value
	log.Infof("[GetPilotSatisTask] total task num is: %v", totalData)

	specialKey := make(map[string]struct{})
	mu := sync.Mutex{}
	signal := make(chan struct{}, 10)
	wg := sync.WaitGroup{}

	for _, hit := range res.Hits.Hits {
		var info UploadMachine
		json.Unmarshal(hit.Source, &info)

		signal <- struct{}{}
		wg.Add(1)

		go func(gInfo UploadMachine) {
			if percentage, ok := CheckTaskPercentValue(gInfo.Machine, metricValue); ok {
				cpuPercent := GetSpecialMetricValue(gInfo.Machine, "cpu")
				ins_ip_key := fmt.Sprintf("%v_%v", gInfo.Machine.InstanceId, gInfo.Machine.InstanceIp)

				mu.Lock()
				if _, recorded := specialKey[ins_ip_key]; recorded {
					mu.Unlock()
					wg.Done()
					<-signal
					return
				}
				mu.Unlock()

				taskIds := getAllTask(gInfo)
				if len(taskIds) == 0 {
					wg.Done()
					<-signal
					return
				}

				if format, taskInfos, sameFormat := CheckTaskInRecording("abc", taskIds); sameFormat {
					log.Infof("[GetPilotSatisTask] instance_type is: %v, instanceTagValue is: %v, instanceId is: %v, instanceIp is: %v, task percentage is: %v, cpu percentage is:  %v, (%v), task number is: %v, %v",
						gInfo.Machine.InstanceType,
						gInfo.TagValue,
						gInfo.Machine.InstanceId,
						gInfo.Machine.InstanceIp,
						percentage,
						cpuPercent,
						format,
						len(taskIds),
						taskInfos,
					)
				} else {
					wg.Done()
					<-signal
					return
				}

				mu.Lock()
				specialKey[ins_ip_key] = struct{}{}
				mu.Unlock()
			}

			wg.Done()
			<-signal
		}(info)
	}

	wg.Wait()

	//for _, hit := range res.Hits.Hits {
	//	if hit.InnerHits != nil {
	//		if innerHits, found := hit.InnerHits["sorted"]; found && len(innerHits.Hits.Hits) > 0 {
	//
	//			for _, item := range innerHits.Hits.Hits {
	//				var info UploadMachine
	//				json.Unmarshal(item.Source, &info)
	//
	//				if CheckTaskPercentValue(info.Machine, metricValue) {
	//					ins_ip_key := fmt.Sprintf("%v_%v", info.Machine.InstanceId, info.Machine.InstanceIp)
	//
	//					if items, ok := allOneInstanceTask[ins_ip_key]; ok {
	//						items = append(items, info)
	//						allOneInstanceTask[ins_ip_key] = items
	//					} else {
	//						allOneInstanceTask[ins_ip_key] = []UploadMachine{info}
	//					}
	//				}
	//			}
	//		}
	//	}
	//}

}

func CheckTaskPercentValue(item InstanceInfoAndTaskInfos, value int) (float64, bool) {
	return checkPercentValue(item, "task", value)
}

func GetSpecialMetricValue(item InstanceInfoAndTaskInfos, name string) float64 {
	for _, metrics := range item.Metrics {
		if metrics.MetricName == name {
			return metrics.MetricValue
		}
	}
	return 0.0
}

func checkPercentValue(item InstanceInfoAndTaskInfos, name string, value int) (float64, bool) {
	if len(item.Metrics) == 0 {
		return 0.0, false
	}

	for _, metrics := range item.Metrics {
		if metrics.MetricName == name && metrics.MetricValue >= float64(value) {
			return metrics.MetricValue, true
		}
	}

	return 0.0, false
}

type PushRecordingMessage struct {
	TaskId          string               `json:"task_id"`
	EventObject     int64                `json:"event_object"`
	ShareMemory     RecordingShareMemory `json:"sharememory"`
	Media           RecordingMedia       `json:"media"`
	Storage         RecordingStorage     `json:"storage"`
	Desc            RecordingDescription `json:"description"`
	InstanceId      string               `json:"instance_id"`
	UpdateTimestamp int64                `json:"update_timestamp"`
	Az              string               `json:"az"`
	DeleteFlag      int                  `json:"delete_flag"`
	BreakingNews    string               `json:"breakingnews"`
	Uid             string               `json:"uid"`
	RequireSign     bool                 `json:"requiresign"`
	CallBackUrl     string               `json:"callback_url"` // 2023-9-11新增
	IsFinal         int                  `json:"is_final"`
}

type RecordingShareMemory struct {
	DecorderPeerId    string `json:"decoder_peer_id"`
	SourcePeerId      string `json:"source_peer_id"`
	SourceType        int    `json:"source_type"`
	SourceName        string `json:"source_name"`
	SourceAddress     string `json:"source_address"`
	PtsOffsetVsSource int    `json:"pts_offset_vs_source"`
	SourceObject      int64  `json:"source_object"`
	SourceBitrate     string `json:"source_bitrate"` // 2022-11-5 新增
	SourceFps         string `json:"source_fps"`     // 2022-11-5 新增
}

type RecordingVCodec struct {
	VInstanceId        string `json:"vinstance_id"`    // 2022-11-5 新增
	VRemoteShmType     string `json:"vremoteshm_type"` // 2022-11-5 新增
	VshmName           string `json:"vshm_name"`
	Resolution         string `json:"resolution"`
	VBitrate           string `json:"vbitrate"`
	ProfileV           string `json:"profileV"`
	GlobalQuality      int    `json:"global_quality"`
	Gop                string `json:"gop"`
	Scale              string `json:"scale"`
	Preset             string `json:"preset"`
	ScanMode           int    `json:"scan_mode"`
	FrameRateScale     int    `json:"frame_rate_scale"`
	FrameRateFrequency int    `json:"frame_rate_frequency"`
	Vcodec             string `json:"vcodec"`
	ExtraRemuxdecode   string `json:"extra_remuxdecode"`  // 2022-11-5 新增
	ExtraFilter        string `json:"extra_filter"`       // 2022-11-5 新增
	ExtraVcodec        string `json:"extra_vcodec"`       // 2022-11-5 新增
	ExtraAcodec        string `json:"extra_acodec"`       // 2022-11-5 新增
	ExtraMux           string `json:"extra_mux"`          // 2022-11-5 新增
	ExtraOutput        string `json:"extra_output"`       // 2022-11-5 新增
	ThumbnailInterval  int    `json:"thumbnail_interval"` // 2023-2-24 新增
	Fps                string `json:"fps"`                // 2023-2-24 新增
}

type RecordingACodec struct {
	AInstanceId string `json:"ainstance_id"` // 2022-11-5 新增
	AshmName    string `json:"ashm_name"`
	Abitrate    string `json:"abitrate"`
	Achannels   int    `json:"achannels"`
	AsampleRate string `json:"asample_rate"`
	Acodec      string `json:"acodec"`
}

type RecordingMeta struct { // 2022-11-5 新增
	MInstanceId       string `json:"minstance_id"`
	MetaShmName       string `json:"metashm_name"`
	MetaRemoteShmAddr string `json:"metaremoteshm_addr"`
}

type RecordingMedia struct {
	FileSize            int64           `json:"file_size"`
	Duration            int64           `json:"duration"`
	TaskStartTimestamp  int64           `json:"task_start_timestamp"` //2023-5-19 新增
	TimecodeStart       int64           `json:"timecode_start"`
	TimecodeEnd         int64           `json:"timecode_end"`
	TimestampStart      int64           `json:"timestamp_start"`
	TimestampEnd        int64           `json:"timestamp_end"`
	RecType             int             `json:"rec_type"`
	VCodec              RecordingVCodec `json:"vcodec"`
	ACodec              RecordingACodec `json:"acodec"`
	Meta                RecordingMeta   `json:"meta"` // 2022-11-5 新增
	SegmentTime         string          `json:"segment_time"`
	PtsOffsetFileSource int             `json:"pts_offset_file_source"`
}

type RecordingStorage struct {
	Type            string `json:"type"`
	Region          string `json:"region"`
	Bucket          string `json:"bucket"`
	RootFolder      string `json:"root_folder"`
	Path            string `json:"path"`
	ThumbnailPath   string `json:"thumbnail_path"`
	MediaObject     int64  `json:"media_object"`
	IndexFullPath   string `json:"index_full_path"`   // 2023-2-24 新增
	FilePathInitial string `json:"file_path_initial"` // 2024-8-8 新增
}

type RecordingDescription struct {
	UserAccount        string `json:"user_account"`
	Project            string `json:"project"`
	Description        string `json:"description"`
	RecordPurpose      string `json:"record_purpose"`
	UserId             string `json:"user_id"`
	ProjectEnv         string `json:"project_env"`
	FileObjectFullPath string `json:"file_object_full_path"` // 2022-2-24新增
	TransferFramType   int    `json:"transfer_frame_type"`   //2023-4-21新增
	ProgramName        string `json:"program_name"`          //2023-4-21新增
	ProgramId          string `json:"program_id"`            //2023-4-21新增
	SlotName           string `json:"slot_name"`             //2023-4-21新增
	SlotId             string `json:"slot_id"`               //2023-4-21新增
	EventName          string `json:"event_name"`            //2023-4-21新增
	EventId            string `json:"event_id"`              //2023-4-21新增
	ExpectedFirstPts   string `json:"expected_first_pts"`    //2023-9-11新增
	ExpectedLastPts    string `json:"expected_last_pts"`     //2023-9-11新增

	SourceObjectId string `json:"source_object_id"` // 2024-6-6 新增
	TangibleId     string `json:"tangible_id"`
}

func CheckTaskInRecording(uniqKey string, tasks []string) (string, string, bool) {
	query := elastic.NewBoolQuery()

	taskQ := elastic.NewTermsQueryFromStrings("task_id", tasks...)
	query.Must(taskQ)

	index := "recording-2024"

	res, err := ESCLI.
		Search().
		Index(index).
		Query(query).
		Size(
			len(tasks),
		).
		Sort("media.timestamp_start", true).
		Do(context.Background())
	if err != nil {
		log.Infof("[CheckTaskInRecording] search from es failed: %v", err)
		return "", "", false
	}

	noEndTimeTask := make([]PushRecordingMessage, 0)
	endTimeTask := make([]PushRecordingMessage, 0)

	for _, hit := range res.Hits.Hits {
		var info PushRecordingMessage
		json.Unmarshal(hit.Source, &info)

		if info.Media.TimestampEnd == 0 {
			noEndTimeTask = append(noEndTimeTask, info)
		} else {
			endTimeTask = append(endTimeTask, info)
		}
	}

	allTaskInfo := ""

	// check normal task
	if len(endTimeTask) > 0 {
		first := endTimeTask[0]
		format := fmt.Sprintf("vbirate: %v, resolution: %v,frame_rate_scale: %v, frame_rate_frequency: %v",
			first.Media.VCodec.VBitrate,
			first.Media.VCodec.Resolution,
			first.Media.VCodec.FrameRateScale,
			first.Media.VCodec.FrameRateFrequency,
		)
		sameFormat := true
		if len(endTimeTask) > 1 {
			for _, item := range endTimeTask {
				allTaskInfo += fmt.Sprintf("task is: %v, project is: %v, global_quality is: %v, extra_filter is: %v;",
					item.TaskId,
					item.Desc.Project,
					item.Media.VCodec.GlobalQuality,
					item.Media.VCodec.ExtraFilter)

				if checkVBitrate(first.Media.VCodec.VBitrate, item.Media.VCodec.VBitrate) &&
					checkFrameRateWild(first.Media.VCodec.FrameRateFrequency, item.Media.VCodec.FrameRateFrequency) &&
					checkGlobalQuality(first.Media.VCodec.GlobalQuality, item.Media.VCodec.GlobalQuality) &&
					checkFrameRateWild(first.Media.VCodec.FrameRateScale, item.Media.VCodec.FrameRateScale) &&
					checkResolution(first.Media.VCodec.Resolution, item.Media.VCodec.Resolution) {
					continue
				}

				sameFormat = false
				break
			}
		} else {
			sameFormat = false
		}

		return format, allTaskInfo, sameFormat

		//if sameFormat {
		//	log.Infof("[CheckTaskInRecording] normal_have_same_format: uniqKey: %v, normal task: %v", uniqKey, marshal(endTimeTask))
		//} else {
		//	log.Infof("[CheckTaskInRecording] normal don't have same format: uniqKey: %v, normal task: %v", uniqKey, marshal(endTimeTask))
		//}
	}
	return "", "", false

	// check un normal task
	//if len(noEndTimeTask) > 0 {
	//	first := noEndTimeTask[0]
	//	sameFormat := true
	//	if len(noEndTimeTask) > 1 {
	//		for _, item := range noEndTimeTask {
	//			if checkVBitrate(first.Media.VCodec.VBitrate, item.Media.VCodec.VBitrate) &&
	//				checkFrameRateWild(first.Media.VCodec.FrameRateFrequency, item.Media.VCodec.FrameRateFrequency) &&
	//				checkFrameRateWild(first.Media.VCodec.FrameRateScale, item.Media.VCodec.FrameRateScale) &&
	//				checkFps(first.Media.VCodec.Fps, item.Media.VCodec.Fps) {
	//				continue
	//			}
	//
	//			sameFormat = false
	//			break
	//		}
	//	} else {
	//		sameFormat = false
	//	}
	//
	//	if sameFormat {
	//		log.Infof("[CheckTaskInRecording] un_normal_have_same_format : uniqKey: %v, has un normal task: %v", uniqKey, marshal(noEndTimeTask))
	//	} else {
	//		log.Infof("[CheckTaskInRecording] un normal don't have same format: uniqKey: %v, has un normal task: %v", uniqKey, marshal(noEndTimeTask))
	//	}
	//}
}

func marshal(data interface{}) string {
	res, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	return string(res)
}

func checkVBitrate(flag string, data string) bool {
	flagNumbers := strings.Split(flag, "K")[0]
	flagNumber, _ := strconv.ParseInt(flagNumbers, 10, 64)

	dataNumbers := strings.Split(data, "K")[0]
	dataNumber, _ := strconv.ParseInt(dataNumbers, 10, 64)

	if flagNumber > dataNumber {
		return flagNumber-dataNumber < 200
	}
	return dataNumber-flagNumber < 200
}

func checkResolution(flag string, data string) bool {
	return flag == data
}

func checkFrameRateWild(flag int, data int) bool {
	if flag > data {
		return flag-data < 5
	}

	return data-flag < 5
}

func checkGlobalQuality(flag int, data int) bool {
	if flag > data {
		return flag-data <= 4
	}

	return data-flag <= 4
}

// fps 只比较逗号以前的
func checkFps(flag string, data string) bool {
	flagPreDot := strings.Split(flag, ".")[0]
	dataPreDot := strings.Split(data, ".")[0]

	return flagPreDot == dataPreDot
}

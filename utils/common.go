package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	log "github.com/cihub/seelog"
	"math/big"
	"os"
)

// UnWrapper 解析多个 es 的参数
func UnWrapper(source []string) []interface{} {
	result := make([]interface{}, len(source))
	for k, v := range source {
		result[k] = v
	}
	return result
}

// ShowDSL 打印 query dsl
func ShowDSL(query interface{}) string {
	queryString := string(GetMarshal(query))
	log.Infof("query is: %v", queryString)
	return queryString
}

// GetMarshal 序列化数据
func GetMarshal(i interface{}) []byte {
	result, err := json.Marshal(i)
	if err != nil {
		log.Errorf("marshal err:%v", err)
		return []byte{}
	}
	return result
}

// GetUnmarshal 反序列化数据
func GetUnmarshal(message string, dest interface{}) error {
	err := json.Unmarshal([]byte(message), &dest)
	if err != nil {
		log.Errorf("unmarshal err: %v", err)
	}
	return err
}

// WriteToPath 写入文件
func WriteToPath(path string, data []string) {
	fileH, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer fileH.Close()
	for _, v := range data {
		fileH.WriteString(v + "\n")
	}
}

// ReadFromPath 读取文件，一行一个string。
func ReadFromPath(path string) []string {
	fileH, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer fileH.Close()

	bufioH := bufio.NewReader(fileH)

	var result []string
	for {
		// 按照 line 分割的。
		line, _, err := bufioH.ReadLine()
		if err != nil {
			break
		}
		result = append(result, string(line))
	}
	return result
}

func BigParse(input string) (string, string) {
	flt, _, err := big.ParseFloat(input, 10, 0, big.ToNearestEven)
	if err != nil {
		log.Debug("[BigParse] parse %v error", input)
		return "", ""
	}
	var i = new(big.Int)
	i, acc := flt.Int(i)
	return fmt.Sprintf("%v", i), fmt.Sprintf("%v", acc)
}

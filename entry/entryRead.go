package entry

import (
	log "github.com/cihub/seelog"
	"io/ioutil"
	"os"
	"strings"
)

func ReadAllEntry(filePath string) []string {
	fileH, err := os.Open(filePath)
	if err != nil {
		log.Errorf("[ReadAllEntry] Open file error: %s", err)
		return nil
	}

	allInfo, err := ioutil.ReadAll(fileH)
	if err != nil {
		log.Errorf("[ReadAllEntry] Read all info error: %s", err)
		return nil
	}

	lines := strings.Split(string(allInfo), "\n")

	// fix the line, to reduce un useful info
	for index, line := range lines {
		if line == "" || strings.Index(line, "#") >= 0 || line == "\r" {
			lines = append(lines[:index], lines[index+1:]...)
		}
	}

	log.Infof("[ReadAllEntry] Read all entry length is: %d", len(lines))

	return lines

}

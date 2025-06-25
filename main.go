package main

import (
	log "github.com/cihub/seelog"
	"wilikidi/es/entry"
	"wilikidi/es/utils"
)

func main() {
	utils.InitLogger()

	utils.InitConfig()

	utils.InitES()

	defer log.Flush()

	filePath := "./entry/allEntry.txt"
	allEntrys := entry.ReadAllEntry(filePath)
	for _, line := range allEntrys {
		log.Infof("all line is: %v", line)
	}
}

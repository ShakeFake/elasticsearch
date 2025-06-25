package entry

import (
	log "github.com/cihub/seelog"
	"testing"
)

func TestReadAllEntry(t *testing.T) {
	path := "./allEntry.txt"

	lines := ReadAllEntry(path)
	for _, line := range lines {
		log.Infof("all line is: %v", line)
	}

}

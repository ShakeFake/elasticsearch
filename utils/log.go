package utils

import log "github.com/cihub/seelog"

var (
	logConfig = `<seelog>
    <outputs formatid="common">
        <rollingfile type="size" filename="./log/log.log" maxsize="10485760" maxrolls="5" />
    </outputs>
    <formats>
        <format id="common" format="%Date %Time %EscM(46)[%LEV]%EscM(49)%EscM(0) [%File:%Line] [%Func] %Msg%n" />
    </formats>
</seelog>`
)

func InitLogger() {
	var logger log.LoggerInterface
	var err error
	logger, err = log.LoggerFromConfigAsFile("./conf/seelog.xml")
	if err != nil {
		log.Infof("init logger failed: %v。 use the default config", err)
		logger, err = log.LoggerFromConfigAsString(logConfig)
	}
	log.ReplaceLogger(logger)
}

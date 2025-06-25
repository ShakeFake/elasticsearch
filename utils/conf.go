package utils

import (
	"fmt"
	"github.com/Unknwon/goconfig"
	log "github.com/cihub/seelog"
)

var (
	ES_HOST   string
	ESUSER    string
	ESPASSWD  string
	ESMAXSIZE int
)

func InitConfig() {
	conf, err := goconfig.LoadConfigFile("./conf/conf.ini")
	if err != nil {
		log.Infof("load config error: %v, use the default online config", err.Error())
		fmt.Printf("load config error: %v, use the default online config\n", err.Error())
		conf = &goconfig.ConfigFile{}
	}

	ES_HOST = conf.MustValue("ES", "es_host", "")
	ESMAXSIZE = conf.MustInt("ES", "es_max_size")
	ESUSER = conf.MustValue("ES", "es_user", "")
	ESPASSWD = conf.MustValue("ES", "es_passwd", "")

	log.Infof("ES_HOST is:%v", ES_HOST)
	log.Infof("ESMAXSIZE is:%v", ESMAXSIZE)
	log.Infof("ESUSER is:%v", ESUSER)
	log.Infof("ESPASSWD is:%v", ESPASSWD)

}

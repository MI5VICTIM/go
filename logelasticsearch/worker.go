package main

import (
	"encoding/json"
	"github.com/millken/go-ipdb"
	"github.com/millken/logger"
	"strings"
	"sync"
)

type LogRecord struct {
	Hostname   string
	Timestamp  string
	Logger     string
	Referer    string `json:"http_referer"`
	Request    string `json:"request"`
	UserAgent  string `json:"http_user_agent"`
	Status     int
	RemoteAddr string `json:"remote_addr"`
	BodySize   int    `json:"body_byte_sent"`
}

var ipDB *ipdb.DB

func loadIpdb() (err error) {
	ipDB, err = ipdb.Load(config.Ipdb)
	return
}

/*
{"index":{"_index":"kangle-2015.01.28","_type":"test.cn"}}
{"Hostname":"cloud.vm","body_bytes_sent":589,"remote_addr":"192.168.3.200","status":200,"http_user_agent":"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.95 Safari/537.36","request":"GET http://www.test.cn/test/ HTTP/1.1","http_referer":"http://www.test.cn/test","Timestamp":"2015-01-28T09:44:44.000Z","Logger":"test.cn"}
*/

func startWorker(i int, ch chan string, wg *sync.WaitGroup) {
	defer wg.Done()

	for _c := range ch {
		go logwork(_c)
		logger.Finest("worker#%d : %s", i, _c)
	}
}

func logwork(logs string) {
	var err error
	var dat map[string]interface{}
	//var dat LogRecord
	log := strings.Split(logs, "\n")

	if len(log) < 2 {
		return
	}
	if err := json.Unmarshal([]byte(log[1]), &dat); err != nil {
		logger.Warn("json Error: %s", err.Error())
		return
	}
	iploc, err := ipDB.Find(dat["remote_addr"].(string))
	if err != nil {
		logger.Warn("ip[%s] Error: %s", dat["remote_addr"], err.Error())
		return
	}
	ip := strings.Split(iploc, "\t")
	dat["country"] = ip[0]
	dat["province"] = ip[1]
	dat["isp"] = strings.Trim(ip[2], "\u0000")

	request := strings.Fields(dat["request"].(string))
	dat["method"] = request[0]
	dat["path"] = strings.SplitN(request[1], "?", 2)[0]

	esCh <- dat
	logger.Debug("%s %v", dat["Hostname"], dat)
}

package main

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"flag"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mediocregopher/radix.v2/redis"
	"github.com/mgutz/str"
	"github.com/sirupsen/logrus"
)

const HANDLE_DIG = " /dig?"
const HANDLE_MOVIE = "/movie/"
const HANDLE_LIST = "/list/"
const HANDLE_HTML = ".html"

type cmdParams struct {
	logFilePath string
	routineNum  int
}

type digData struct {
	time  string
	url   string
	refer string
	ua    string
}

type urlData struct {
	data  digData
	uid   string
	unode urlNode
}

type urlNode struct {
	unType string
	unRid  int
	unUrl  string
	unTime string
}

type storageBlock struct {
	counterType  string
	storageModel string
	unode        urlNode
}

var log = logrus.New()
var redisClient redis.Client

func init() {
	log.Out = os.Stdout
	log.SetLevel(logrus.DebugLevel)
	redisClient, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		log.Fatalln("Radis connect failed")
	} else {
		defer redisClient.Close()
	}
}

func main() {
	// read command line parameter
	// usage:
	//    go run run.go --routineNum=5 --logFilePath=/thepath
	logFilePath := flag.String("logFilePath", "/User/", "log file path")
	routineNum := flag.Int("routineNum", 5, "how many consumer goroutine")
	l := flag.String("l", "/temp/log", "this app runtime log target file path")
	flag.Parse()

	params := cmdParams{
		*logFilePath, *routineNum,
	}
	// write log
	logFd, err := os.OpenFile(*l, os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		log.Out = logFd
		defer logFd.Close()
	}
	log.Infoln("Exec start.")
	// channels
	var logChannel = make(chan string, params.routineNum*3)
	var pvChannel = make(chan urlData, params.routineNum)
	var uvChannel = make(chan urlData, params.routineNum)
	var storageChannel = make(chan storageBlock, params.routineNum)

	go readFileLineByLine(params, logChannel)

	for i := 0; i < params.routineNum; i++ {
		go logConsumer(logChannel, pvChannel, uvChannel)
	}

	go pvCounter(pvChannel, storageChannel)
	go uvCounter(uvChannel, storageChannel)

	go dataStorage(storageChannel)

	//
	time.Sleep(1000 * time.Second)
}

/*
	Read log line by line and put log line to the log channel
*/
func readFileLineByLine(params cmdParams, logChannel chan string) error {
	fd, err := os.Open(params.logFilePath)
	if err != nil {
		log.Warningf("Can not open file:%s", params.logFilePath)
		return err
	}

	defer fd.Close()
	count := 0
	buffereRead := bufio.NewReader(fd)
	for {
		line, err := buffereRead.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(3 * time.Second)
				log.Info("ReadFileLineByLine wait")
			} else {
				log.Info("ReadFileLineByLine read log error")
			}
		}
		logChannel <- line
		count++

		if count%(1000*params.routineNum) == 0 {
			log.Info("ReadFileLineByLine line: %d", count)
		}
	}
}

/*
	N log consumer take log line from log channel and parse it
	Put parse result into pv channel or uv channel
*/
func logConsumer(logChannel chan string, pvChannel, uvChannel chan urlData) error {
	for logStr := range logChannel {
		data := cutLogFetchData(logStr)

		// uid = m5(refer+ua)
		hasher := md5.New()
		hasher.Write([]byte(data.refer + data.ua))
		uid := hex.EncodeToString(hasher.Sum(nil))
		uData := urlData{
			data,
			uid,
			formatUrl(data.url, data.time),
		}

		pvChannel <- uData
		uvChannel <- uData
	}
	return nil
}

func formatUrl(url, t string) urlNode {
	// 一定从量大的着手,  详情页>列表页≥首页
	pos1 := str.IndexOf(url, HANDLE_MOVIE, 0)
	if pos1 != -1 {
		pos1 += len(HANDLE_MOVIE)
		pos2 := str.IndexOf(url, HANDLE_HTML, 0)
		idStr := str.Substr(url, pos1, pos2-pos1)
		id, _ := strconv.Atoi(idStr)
		return urlNode{"movie", id, url, t}
	} else {
		pos1 = str.IndexOf(url, HANDLE_LIST, 0)
		if pos1 != -1 {
			pos1 += len(HANDLE_LIST)
			pos2 := str.IndexOf(url, HANDLE_HTML, 0)
			idStr := str.Substr(url, pos1, pos2-pos1)
			id, _ := strconv.Atoi(idStr)
			return urlNode{"list", id, url, t}
		} else {
			return urlNode{"home", 1, url, t}
		} // 如果页面url有很多种，就不断在这里扩展
	}
}

func cutLogFetchData(logStr string) digData {
	logStr = strings.TrimSpace(logStr)
	pos1 := str.IndexOf(logStr, HANDLE_DIG, 0)
	if pos1 == -1 {
		return digData{}
	}

	pos1 += len(HANDLE_DIG)

	pos2 := str.IndexOf(logStr, " HTTP/", pos1)
	d := str.Substr(logStr, pos1, pos2-pos1) // paramters in url

	urlInfo, err := url.Parse("http://localhost/?" + d)
	if err != nil {
		return digData{}
	}
	data := urlInfo.Query()
	return digData{
		data.Get("time"),
		data.Get("refer"),
		data.Get("url"),
		data.Get("ua"),
	}
}

func pvCounter(pvChannel chan urlData, storageChannel chan storageBlock) {
	for data := range pvChannel {
		sItem := storageBlock{"pv", "ZINCREBY", data.unode}
		storageChannel <- sItem
	}
}

func uvCounter(uvChannel chan urlData, storageChannel chan storageBlock) {
	for data := range uvChannel {
		//HyperLoglog redis
		hyperLogLogKey := "uv_hpll_" + getTime(data.data.time, "day")
		ret, err := redisClient.Cmd("PFADD", hyperLogLogKey, data.uid, "EX", 86400).Int()
		if err != nil {
			log.Warningln("UvCounter check redis hyperloglog failed, ", err)
		}
		if ret != 1 {
			continue
		}

		sItem := storageBlock{"uv", "ZINCRBY", data.unode}
		storageChannel <- sItem
	}
}

func dataStorage(storageChannel chan storageBlock) {
	for block := range storageChannel {
		prefix := block.counterType + "_"

		// 逐层添加，加洋葱皮的过程
		// 维度： 天-小时-分钟
		// 层级： 定级-大分类-小分类-终极页面
		// 存储模型： Redis  SortedSet
		setKeys := []string{
			prefix + "day_" + getTime(block.unode.unTime, "day"),
			prefix + "hour_" + getTime(block.unode.unTime, "hour"),
			prefix + "min_" + getTime(block.unode.unTime, "min"),
			prefix + block.unode.unType + "_day_" + getTime(block.unode.unTime, "day"),
			prefix + block.unode.unType + "_hour_" + getTime(block.unode.unTime, "hour"),
			prefix + block.unode.unType + "_min_" + getTime(block.unode.unTime, "min"),
		}

		rowId := block.unode.unRid

		for _, key := range setKeys {
			ret, err := redisPool.Cmd(block.storageModel, key, 1, rowId).Int()
			if ret <= 0 || err != nil {
				log.Errorln("DataStorage redis storage error.", block.storageModel, key, rowId)
			}
		}
	}
}

func getTime(logTime, timeType string) string {
	var item string
	switch timeType {
	case "day":
		item = "2006-01-02"
		break
	case "hour":
		item = "2006-01-02 15"
		break
	case "min":
		item = "2006-01-02 15:04"
		break
	}
	t, _ := time.Parse(item, time.Now().Format(item))
	return strconv.FormatInt(t.Unix(), 10)
}

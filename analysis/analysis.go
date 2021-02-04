package main

import (
	"flag"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

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
	data digData
	uid  string
}

type urlNode struct {
}

type storageBlock struct {
	counterType  string
	storageModel string
	unode        urlNode
}

var log = logrus.New()

func init() {
	log.Out = os.Stdout
	log.SetLevel(logrus.DebugLevel)
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

func readFileLineByLine(params cmdParams, logChannel chan string) {
	fd, err := os.OpenFile(params.logFilePath)
	if err != nil {
		log.Warningln()
	}
}

func logConsumer(logChannel chan string, pvChannel, uvChannel chan urlData) {

}

func pvCounter(pvChannel chan urlData, storageChannel chan storageBlock) {

}
func uvCounter(pvChannel chan urlData, storageChannel chan storageBlock) {

}

func dataStorage(storageChannel chan storageBlock) {

}

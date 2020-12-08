package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	selfcheck "github.com/gangjun06/auto-selfcheck"
)

type List struct {
	School string `json:"school"`
	Name   string `json:"name"`
	Birth  string `json:"birth"`
}

type Log struct {
	File *os.File
}

func NewLog(path *string) *Log {
	f, err := os.OpenFile(*path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	return &Log{File: f}
}

func (l *Log) AppendLog(text string) {
	l.File.WriteString(time.Now().Format(time.RFC3339) + ": (info) " + text + "\n")
}

func (l *Log) AppendError(text string) {
	l.File.WriteString(time.Now().Format(time.RFC3339) + ": (error) " + text + "\n")
}

func main() {
	rand.Seed(time.Now().UnixNano())

	logFilePath := flag.String("log", "", "log file path")
	listPath := flag.String("list", "./list.json", "List of self-check participants")
	delayTime := flag.Int("delay", 300, "delay (second)")

	flag.Parse()

	f := NewLog(logFilePath)
	f.AppendLog("-----------------------")
	f.AppendLog(time.Now().Format("2006년 01월 02일 15시 04분 05초"))
	f.AppendLog("자동 자가진단 시작")

	jsonFile, err := ioutil.ReadFile(*listPath)
	if err != nil {
		f.AppendError(err.Error())
		log.Fatalln(err)
		return
	}

	var data []List
	if err := json.Unmarshal(jsonFile, &data); err != nil {
		f.AppendError(err.Error())
		log.Fatalln(err)
		return
	}

	rand.Shuffle(len(data), func(i, j int) { data[i], data[j] = data[j], data[i] })

	for _, d := range data {
		if *delayTime != 0 {
			n := rand.Intn(*delayTime)
			time.Sleep(time.Duration(n) * time.Second)
		}
		info, err := selfcheck.GetStudnetInfo(selfcheck.AREA_GYEONGBUK, d.School, d.Name, d.Birth)
		if err != nil {
			log.Println(err)
			f.AppendError(fmt.Sprintf("%s의 학생정보 불러오기 [실패]", d.Name))
			continue
		}
		if err := info.AllHealthy(); err != nil {
			log.Println(err)
			f.AppendError(fmt.Sprintf("%s의 자가진단 참여 [실패]", d.Name))
			continue
		}
		f.AppendError(fmt.Sprintf("%s의 자가진단 참여 [성공]", d.Name))
	}
}

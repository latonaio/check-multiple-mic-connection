package cmd

import (
	"github.com/mattn/go-pipeline"
	"log"
	"strconv"
	"strings"
)

type MicInfo struct {
	CardNo   int
	DeviceNo int
}

func MicrophoneList() []*MicInfo {
	list := getAllMicrophoneFromAlsa()
	if len(list) == 0 {
		return nil
	}
	micList := make([]*MicInfo, len(list))
	for i, v := range list {
		micList[i] = parseAlsaMicInfo(v)
	}
	return micList
}

func getAllMicrophoneFromAlsa() []string {
	out, err := pipeline.Output(
		[]string{"arecord", "-l"},
		[]string{"grep", "Microphone"},
	)
	if err != nil {
		log.Println(err)
	}
	res := strings.Split(string(out), "\n")
	return res[:len(res)-1]
}

func parseAlsaMicInfo(info string) *MicInfo {
	splited := strings.Split(info, " ")
	cardNo, _ := strconv.Atoi(splited[1][0:1])
	deviceNo, _ := strconv.Atoi(splited[6][0:1])
	return &MicInfo{
		CardNo:   cardNo,
		DeviceNo: deviceNo,
	}
}

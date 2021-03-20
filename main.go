package main

import (
	"check-multiple-mic-connection/cmd"
	"check-multiple-mic-connection/config"
	db "check-multiple-mic-connection/pkg/db"
	"context"
	"errors"
	"github.com/jinzhu/gorm"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	aion_log "bitbucket.org/latonaio/aion-core/pkg/log"
)

var statusProcessList = map[int]bool{}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}
	mysql := db.MysqlDB{}
	mdb := mysql.Connect(cfg)
	// DB内のマイク接続情報を全て削除して初期化する
	err = cmd.InitMicStatus(mdb.GetConn())
	if err != nil {
		aion_log.Printf("failed to init mic status. err = %s", err)
	}

	errCh := make(chan error, 1)
	quitC := make(chan os.Signal, 1)
	signal.Notify(quitC, syscall.SIGTERM, os.Interrupt)

	err = cmd.NewKanbanClient(ctx)
	if err != nil {
		aion_log.Printf("failed to init kanban client. err = %s", err)
	}
	err = cmd.SetKanban()
	if err != nil {
		aion_log.Printf("failed to set kanban. err = %s", err)
	}

	go watch(ctx, time.Second*5, errCh)

loop:
	for {
		select {
		case err := <-errCh:
			log.Println(err)
			break loop
		case q := <-quitC:
			aion_log.Print("stop")
			err := cmd.KanbanCloseConn()
			if err != nil {
				errCh <- err
			}
			log.Printf("signal received. %s", q.String())
			time.Sleep(5 * time.Second)
			break loop
		}
	}

}

func watch(ctx context.Context, interval time.Duration, errCh chan error) {
	log.Println("start watch!")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	err := manageMicConn()
	if err != nil {
		errCh <- err
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err = manageMicConn()
			if err != nil {
				errCh <- err
			}
		}
	}
}

func manageMicConn() error {
	dbConn := db.GetMysql().GetConn()
	var m sync.Mutex
	m.Lock()
	defer m.Unlock()
	alsa := cmd.MicrophoneList()
	if alsa == nil {
		log.Println("no microphone found.")
		return nil
	}

	allMic, err := cmd.GetAllMicrophones(db.GetMysql().GetConn())
	if err != nil {
		return err
	}

	deleteMic := make([]*cmd.Microphone, len(allMic))
	copy(deleteMic, allMic)

	for i := 0; i < len(deleteMic); i++ {
		for _, al := range alsa {
			if deleteMic[0].CardNo == al.CardNo && deleteMic[0].DeviceNo == al.DeviceNo {
				deleteMic = append(deleteMic[:i], deleteMic[i+1:]...)
			}
		}
	}
	log.Printf("check disconnected mic complete. delete list is %+v",deleteMic)
	log.Printf("statusProcessList is %+v",statusProcessList)
	for _, v := range deleteMic {
		if v == nil {
			continue
		}
		err := v.UpdateStatus(cmd.DISABLE,dbConn)
		if err != nil {
			return err
		}
		statusProcessList[v.ManagerPodProcessNum] = false
	}
	log.Printf("delete mic complete")

	log.Printf("%+v", alsa)
	// マイクの接続処理を行う
	// 一度マイクを接続すると、DBにalsaの割り当て記録がサウンドカード番号とデバイス番号とともに記録される
	// 切断を検知するとstatusがdisableとなるので、disableの割り当て情報があれば、そこに優先的にマイクの再割り当てを行う
	// 既存の割り当て情報が全て埋まっていた場合、新規で割り当てを行う
	for i, v := range alsa {
		pNum := getAvailableProcessNum(i + 1)
		m, err := cmd.GetMicByCardNoAndDevNo(v.CardNo, v.DeviceNo, dbConn)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("Fetch error %v", err)
				return err
			}
			log.Printf("Create new mic record %v", v)
			err := cmd.InsertMicrophone(v.CardNo, v.DeviceNo, dbConn)
			if err != nil {
				log.Printf("insert error %v", err)
				return err
			}
		}
		if m != nil {
			log.Printf("%+v", m)
			if m.Status == cmd.ACTIVE {
				continue
			}
			if m.Status == cmd.DISABLE {
				pNum = m.ManagerPodProcessNum
				err := m.UpdateStatus(cmd.STANDBY,dbConn)
				if err != nil {
					return err
				}
			}
		}

		log.Printf("start capture audio,%v,%v,%v",v.CardNo, v.DeviceNo, pNum)
		err = StartCaptureAudioService(v.CardNo, v.DeviceNo, pNum)
		if err != nil {
			return err
		}
		statusProcessList[pNum] = true
	}
	return nil
}

func StartCaptureAudioService(cardNo, deviceNo, order int) error {
	reqData := map[string]interface{}{
		"card_no":   cardNo,
		"device_no": deviceNo,
		"status":    2,
	}
	if err := cmd.WriteKanban(reqData, order, "default"); err != nil {
		return err
	}
	return nil
}

// マイクの抜き差しによって虫食い状にprocessNumの空きが発生するので、ここで空き番を取得する
func getAvailableProcessNum(idx int) int {
	if len(statusProcessList) == 0 || len(statusProcessList) < idx {
		return idx
	}

	for k, v := range statusProcessList {
		if !v {
			return k
		}
	}
	return idx
}

package main

import (
	"check-multiple-mic-connection/cmd"
	"check-multiple-mic-connection/config"
	db "check-multiple-mic-connection/pkg/db"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	aion_log "bitbucket.org/latonaio/aion-core/pkg/log"
)

/*
端末へのマイクの接続情報を確認する
dbのレコードの数とalsaのinfoの差分を確認する
alsa > db なら差分だけcapture-audio-from-micを立ち上げる
alsa < db ならcardnoとdevicenoで引いたpodを落とす（？）<-　一旦いらないかも
差分の端末の数ぶんcapture-audio-from-micを立ち上げるようkanbanに配信
capture-audio-from-micがdbの接続者情報に自分のpod名を入れる
podが死んでる時の挙動は一旦考えない
 */

func main()  {
	ctx, cancel := context.WithCancel(context.Background())
	cfg,err := config.New()
	if err != nil {
		cancel()
		panic(err)
	}
	mysql := db.MysqlDB{}
	mdb := mysql.Connect(cfg)
	// DB内のマイク接続情報を全て削除して初期化する
	err = cmd.InitMicStatus(mdb.GetConn())
	if err != nil {
		aion_log.Printf("failed to init mic status. err = %s",err)
	}

	errCh := make(chan error,1)
	quitC := make(chan os.Signal, 1)
	signal.Notify(quitC, syscall.SIGTERM, os.Interrupt)

	err = cmd.NewKanbanClient(ctx)
	if err != nil {
		aion_log.Printf("failed to init kanban client. err = %s",err)
	}
	err = cmd.SetKanban()
	if err != nil {
		aion_log.Printf("failed to set kanban. err = %s",err)
	}


	go watch(ctx,time.Second*5,errCh)


	select {
	case err := <-errCh:
		log.Println(err)
	case <-quitC:
		aion_log.Print("stop")
		err := cmd.KanbanCloseConn()
		if err != nil {
			errCh <- err
		}
		time.Sleep(time.Second*5)
		cancel()
	}
}

func watch(ctx context.Context, interval time.Duration,errCh chan error) {
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
			err =  manageMicConn()
			if err != nil {
				errCh <- err
			}
		}
	}
}

func manageMicConn() error {
	dbConn := db.GetMysql().GetConn()
	alsa := cmd.MicrophoneList()
	if alsa == nil {
		log.Println("no microphone found.")
		return nil
	}
	for i,v := range alsa {
		// 未接続のマイクの接続が検出された時だけrecordの挿入とpodの立ち上げを行う
		if !cmd.CheckMicrophoneExists(v.CardNo,v.DeviceNo,dbConn) {
			log.Printf("new microphone is detected. cardNo=%d,deviceNo=%d",v.CardNo,v.DeviceNo)
			err := cmd.InsertMicrophoneIfNotExist(v.CardNo,v.DeviceNo,dbConn)
			if err != nil {
				return err
			}
			err = StartCaptureAudioService(v.CardNo,v.DeviceNo,i+1)
			if err != nil {
				return err
			}
		}
	}
	return nil
}


func StartCaptureAudioService(cardNo,deviceNo,order int) error {
	reqData := map[string]interface{}{
		"card_no":cardNo,
		"device_no": deviceNo,
	}
	if err := cmd.WriteKanban(reqData,order); err != nil {
		return err
	}
	return nil
}
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
	mysql.Connect(cfg)

	errCh := make(chan error,1)
	quitC := make(chan os.Signal, 1)
	signal.Notify(quitC, syscall.SIGTERM, os.Interrupt)

	go watch(ctx,time.Second*5,errCh)


	select {
	case err := <-errCh:
		log.Println(err)
	case <-quitC:
		log.Print("stop")
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
	log.Printf("alsa %+v",alsa)
	for _,v := range alsa {
		log.Println("looping")
		err := cmd.InsertMicrophoneIfNotExist(v.CardNo,v.DeviceNo,dbConn)
		if err != nil {
			return err
		}
		err = StartCaptureAudioService()
		if err != nil {
			return err
		}
	}
	return nil
}


func StartCaptureAudioService() error {
	log.Println("Start Microservice")
	return nil
}
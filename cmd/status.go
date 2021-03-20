package cmd

import (
	"github.com/jinzhu/gorm"
)

// TODO 増えそうならenumに
const (
	STANDBY = "standby"
	ACTIVE  = "active"
	DISABLE = "disable"
)

type Microphone struct {
	CardNo               int
	DeviceNo             int
	Status               string
	ManagerPodProcessNum int
}

func (m *Microphone) InsertMicrophone(db *gorm.DB) error {
	return db.Create(&m).Error
}

func GetAllMicrophones(db *gorm.DB) ([]*Microphone, error) {
	microphones := []*Microphone{}
	if err := db.Find(&microphones).Error; err != nil {
		return nil, err
	}
	return microphones, nil
}

func InsertMicrophone(cardNo, deviceNo int, db *gorm.DB) error {
	return db.Create(&Microphone{
		CardNo:   cardNo,
		DeviceNo: deviceNo,
		Status:   STANDBY,
	}).Error
}

func CheckMicrophoneExists(cardNo, deviceNo int, db *gorm.DB) bool {
	var cnt int
	db.Model(&Microphone{}).Where("card_no = ? AND device_no = ?", cardNo, deviceNo).Count(&cnt)
	return cnt > 0
}

func GetDisableMicrophone( db *gorm.DB) []*Microphone {
	mics := []*Microphone{}
	db.Where("status = ?", ACTIVE).Find(&mics)

	return mics
}

func GetMicByCardNoAndDevNo(cardNo, deviceNo int, db *gorm.DB) (*Microphone,error) {
	m := &Microphone{}
	err := db.Where("card_no = ? AND device_no = ?", cardNo, deviceNo).First(&m).Error
	return m,err
}

func InitMicStatus(db *gorm.DB) error {
	return db.Delete(&Microphone{}).Error
}

func (m *Microphone) UpdateStatus(status string,db *gorm.DB) error {
	return db.Model(&Microphone{}).Where("card_no = ? AND device_no = ?", m.CardNo, m.DeviceNo).Update("status", status).Error
}

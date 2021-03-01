package cmd

import "github.com/jinzhu/gorm"

type Microphone struct {
	CardNo int
	DeviceNo int
	AvailableFlg bool
}

func (m *Microphone)InsertMicrophone(db *gorm.DB) error {
	return db.Create(&m).Error
}

func GetAllMicrophones(db *gorm.DB) ([]*Microphone,error) {
	microphones := []*Microphone{}
	if err := db.Find(&microphones).Error; err != nil {
		return nil,err
	}
	return microphones,nil
}

func InsertMicrophoneIfNotExist(cardNo,deviceNo int,db *gorm.DB) error {
	m := &Microphone{}
	return db.Where("card_no = ? AND device_no = ?", cardNo,deviceNo).Assign(&Microphone{
		CardNo: cardNo,
		DeviceNo:deviceNo,
	}).FirstOrCreate(&m).Error
}

func CheckMicrophoneExists(cardNo,deviceNo int,db *gorm.DB) bool {
	var cnt int
	db.Model(&Microphone{}).Where("card_no = ? AND device_no = ?", cardNo,deviceNo).Count(&cnt)
	return cnt > 0
}

func InitMicStatus(db *gorm.DB) error {
	return db.Delete(&Microphone{}).Error
}
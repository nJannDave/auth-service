package entity

import "errors"

type UserData struct {
	AccountId int `gorm:"column:account_id;autoIncrement"`
	NIK string `gorm:"column:nik"`
	Name string `gorm:"column:name"`
}

func (UserData) TableName() string {
	return "account"
}

func (u *UserData) ValidateNik() error {
	if len(u.NIK) != 16 {
		return errors.New("nik must be 16 digits")
	}
	return nil
}

func NewUserData(nik string, name string) *UserData {
	return &UserData{
		NIK: nik,
		Name: name,
	}
}

type Residence struct {
	Province string `gorm:"column:province"`
	City string `gorm:"column:city"`
}

func NewResidence(province string, city string) *Residence {
	return &Residence{
		Province: province,
		City: city,
	}
}

type JunctionData struct {
	ProvinceId int `gorm:"column:province_id"`
	AccountId int `gorm:"column:account_id"`
	CityId int `gorm:"column:city_id"`
}

func NewJuctionData(pid int, aid int, cid int) *JunctionData {
	return &JunctionData{
		ProvinceId: pid,
		AccountId: aid,
		CityId: cid,
	}
}

func (JunctionData) TableName() string {
	return "junction"
}
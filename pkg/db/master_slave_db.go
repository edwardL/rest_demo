package db

import "gorm.io/gorm"

type MsConfig struct {
	Master Config
	Slave  Config
}

type MsDb struct {
	master *gorm.DB
	slave  *gorm.DB
}

func NewMsDB(conf *MsConfig) (_ *MsDb, err error) {
	ms := &MsDb{}
	if ms.master, err = NewDB(conf.Master); err != nil {
		return nil, err
	}

	if ms.slave, err = NewDB(conf.Slave); err != nil {
		return nil, err
	}
	return ms, nil
}

func (mdb *MsDb) Master() *gorm.DB {
	return mdb.master
}

func (mdb *MsDb) Slave() *gorm.DB {
	return mdb.slave
}

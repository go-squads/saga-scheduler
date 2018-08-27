package main

import (
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type lxcService struct {
	ID      string `db:"id" json:"id"`
	Service string `db:"service" json:"service"`
	LxcID   string `db:"lxc_id" json:"lxc_id"`
	LxcPort string `db:"lxc_port" json:"lxc_port"`
	LxdID   string `db:"lxd_id" json:"lxd_id"`
	LxdPort string `db:"lxd_port" json:"lxd_port"`
	LxcName string `db:"lxc_name" json:"lxc_name"`
	Status  string `db:"status" json:"status"`
	LxdName string `db:"lxd_name" json:"lxd_name"`
}

func (l *lxcService) insertLxcService(db *sqlx.DB) error {
	_, err := db.NamedExec(`INSERT INTO lxc_services (id, service, lxc_id, lxc_port, lxd_id, lxd_port, lxc_name, status, lxd_name) VALUES (:id, :service, :lxc_id, :lxc_port, :lxd_id, :lxd_port, :lxc_name, 'creating', :lxd_name)`, l)
	if err != nil {
		return err
	}

	return nil
}

func (l *lxcService) checkIfLxcServiceExist(db *sqlx.DB) bool {
	rows, err := db.Queryx("SELECT id, service, lxc_id, lxc_port, lxd_id, lxd_port FROM lxc_services WHERE lxc_port=$1 AND lxd_port=$2", l.LxcPort, l.LxdPort)
	if err != nil {
		return false
	}
	defer rows.Close()

	if rows.Next() {
		return true
	}
	return false
}

func getLxcServicesList(db *sqlx.DB, lxdName string) ([]lxcService, error) {
	rows, err := db.Queryx("SELECT id, service, lxc_id, lxc_port, lxd_id, lxd_port, lxc_name, status FROM lxc_services WHERE lxd_name=$1", lxdName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var lxcServiceList []lxcService
	for rows.Next() {
		l := lxcService{}
		if err = rows.StructScan(&l); err != nil {
			return nil, err
		}
		lxcServiceList = append(lxcServiceList, l)
	}

	return lxcServiceList, nil
}

func (l *lxcService) getLxcService(db *sqlx.DB) error {
	rows, err := db.Queryx("SELECT id, service, lxc_id, lxc_port, lxd_id, lxd_port, lxc_name, status FROM lxc_services WHERE id=$1", l.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.StructScan(&l)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *lxcService) checkNeedUpdateLxcService(curLxc lxcService) bool {
	if curLxc.Status == "created" {
		if l.Status == "creating" {
			return false
		} else {
			return true
		}
	}

	return true
}

func (l *lxcService) updateStatusByID(db *sqlx.DB) error {
	curLxc := lxcService{ID: l.ID}
	if err := curLxc.getLxcService(db); err != nil {
		return err
	}
	if l.checkNeedUpdateLxcService(curLxc) {
		log.Info("Lxc status update needed")
		_, err := db.Exec("UPDATE lxc_services SET status = $2 WHERE id = $1", l.ID, l.Status)
		if err != nil {
			return err
		}
	}
	return nil
}

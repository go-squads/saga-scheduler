package main

import (
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type lxc struct {
	ID          string `db:"id" json:"id"`
	LxdID       string `db:"lxd_id" json:"lxd_id"`
	Name        string `db:"name" json:"name"`
	Type        string `db:"type" json:"type"`
	Alias       string `db:"alias" json:"alias"`
	Protocol    string `db:"protocol" json:"protocol"`
	Server      string `db:"server" json:"server"`
	Address     string `db:"address" json:"address"`
	Status      string `db:"status" json:"status"`
	Description string `db:"description" json:"description"`
}

func (l *lxc) getLxc(db *sqlx.DB) error {
	rows, err := db.Queryx("SELECT id, lxd_id, name, type, alias, address, description, status FROM lxc WHERE id=$1 LIMIT 1", l.ID)
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

func (l *lxc) checkNeedUpdate(curLxc lxc) bool {
	if curLxc.Status == "started" {
		if l.Status == "starting" {
			return false
		} else {
			return true
		}
	} else if curLxc.Status == "stopped" {
		if l.Status == "stopping" {
			return false
		} else {
			return true
		}
	}
	return true
}

func (l *lxc) updateStatusByID(db *sqlx.DB) error {
	curLxc := lxc{ID: l.ID}
	if err := curLxc.getLxc(db); err != nil {
		return err
	}
	if l.checkNeedUpdate(curLxc) {
		log.Info("Lxc status update needed")
		_, err := db.Exec("UPDATE lxc SET status = $2 WHERE id = $1", l.ID, l.Status)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *lxc) insertLxc(db *sqlx.DB) error {
	_, err := db.NamedExec("INSERT INTO lxc (id, lxd_id, name, type, alias, protocol, server, address, description, status) VALUES (:id, :lxd_id, :name, :type, :alias, :protocol, :server, :address, :description, :status)", l)
	if err != nil {
		return err
	}

	return nil
}

func (l *lxc) deleteLxc(db *sqlx.DB) error {
	_, err := db.Queryx("DELETE FROM lxc WHERE id = $1", l.ID)
	if err != nil {
		return err
	}
	return nil
}

func (l *lxc) getLxcListByLxdID(db *sqlx.DB, lxdID string) ([]lxc, error) {
	rows, err := db.Queryx("SELECT id, lxd_id, name, type, alias, protocol, server, address, description, status FROM lxc WHERE lxd_id=$1", lxdID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lxcList := []lxc{}

	for rows.Next() {
		l := lxc{}
		err = rows.StructScan(&l)
		if err != nil {
			return nil, err
		}
		lxcList = append(lxcList, l)
	}

	return lxcList, nil
}

package main

import (
	log "github.com/sirupsen/logrus"
)

type lxd struct {
	ID          string `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Address     string `db:"address" json:"address"`
	Description string `db:"description" json:"description"`
}

func (l *lxd) getLxd(db PostgresQL) error {
	rows, err := db.Queryx("SELECT id, name, address, description FROM lxd WHERE id=$1 LIMIT 1", l.ID)
	if err != nil {
		return err
	}

	if rows.Next() {
		err = rows.StructScan(&l)
		if err != nil {
			return err
		}
	}
	return nil
}

func getLxds(db PostgresQL) ([]lxd, error) {
	var result []lxd
	rows, err := db.Queryx("SELECT id, name, address, description FROM lxd")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var temp lxd
		err = rows.StructScan(&temp)
		if err != nil {
			return nil, err
		}
		result = append(result, temp)
	}
	return result, nil
}

func (l *lxd) getLxdByIP(db PostgresQL) error {
	rows, err := db.Queryx("SELECT id, address FROM lxd WHERE address=$1 LIMIT 1", l.Address)
	if err != nil {
		return err
	}

	if rows.Next() {
		err = rows.StructScan(&l)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *lxd) insertLxd(db PostgresQL) error {
	_, err := db.NamedExec("INSERT INTO lxd (id, name, address, description) VALUES (:id, :name, :address, :description)", l)
	if err != nil {
		return err
	}

	return nil
}

func (l *lxd) getLxdIDByName(db PostgresQL) error {
	rows, err := db.Queryx("SELECT id, name, address, description FROM lxd WHERE name = $1", l.Name)
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

func (l *lxd) getLxdNameAndAddressByID(db PostgresQL) (string, string) {
	rows, err := db.Queryx("SELECT name, address FROM lxd WHERE id=$1", l.ID)
	if err != nil {
		log.Error(err.Error())
		return "", ""
	}
	defer rows.Close()

	lxdData := lxd{}
	if rows.Next() {
		if err = rows.StructScan(&lxdData); err != nil {
			log.Error(err.Error())
			return "", ""
		}
	}
	log.Infof("lxd name: %s", lxdData.Name)
	return lxdData.Address, lxdData.Name
}

package main

import (
	"github.com/jmoiron/sqlx"
)

type operation struct {
	ID          string `db:"id" json:"id"`
	LxcID       string `db:"lxc_id" json:"lxc_id"`
	Status      string `db:"status" json:"status"`
	StatusCode  int    `db:"status_code" json:"status_code"`
	Description string `db:"description" json:"description"`
}

func (op *operation) getOperation(db *sqlx.DB) error {
	rows, err := db.Queryx("SELECT * FROM operation WHERE id=$1 LIMIT 1", op.ID)
	if err != nil {
		return err
	}

	if rows.Next() {
		err = rows.StructScan(&op)
		if err != nil {
			return err
		}
	}
	return nil
}

func (op *operation) insertOperation(db *sqlx.DB) error {
	_, err := db.NamedExec("INSERT INTO operation (id, lxc_id, status, status_code) VALUES (:id, :lxc_id, :status, :status_code)", op)
	if err != nil {
		return err
	}

	return nil
}

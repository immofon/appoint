package main

import (
	"fmt"
	"time"

	"github.com/immofon/appoint"
	"go.etcd.io/bbolt"
)

func main() {
	db, _ := bbolt.Open("tmp.bolt", 0x600, &bbolt.Options{
		Timeout: time.Second * 1,
	})
	defer db.Close()

	appoint.Prepare(db)
	db.Update(func(tx *bbolt.Tx) error {
		appoint.SetRole(tx, "1", appoint.Role_Admin)
		appoint.SetRole(tx, "2", appoint.Role_Teacher)
		for id := 3; id <= 254; id++ {
			appoint.SetRole(tx, fmt.Sprint(id), appoint.Role_Student)
		}
		return nil
	})
}

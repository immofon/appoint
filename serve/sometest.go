package main

import (
	"fmt"

	"github.com/immofon/appoint"
	bolt "go.etcd.io/bbolt"
)

func Sometest(db *bolt.DB) {
	db.View(func(tx *bolt.Tx) error {
		appoint.UpdateData(tx)

		return nil
	})

	fmt.Println(appoint.GetData())
}

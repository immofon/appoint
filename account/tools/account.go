package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/immofon/appoint/account"
	"github.com/immofon/appoint/log"
	bolt "go.etcd.io/bbolt"
)

func main() {
	log.TextMode()
	db, err := bolt.Open("tmp.bolt", 0x600, &bolt.Options{
		Timeout: time.Second * 1,
	})

	NoErr(err)

	account.Prepare(db)
	list(db)
	account.Auth(db, "moon", "1223")
	add(db, "moon", "123", "cxy")
	get(db, "1")
	get(db, "moon")
	add(db, "moon", "123", "cxy")
}

func NoErr(err error) {
	return
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func add(db *bolt.DB, _account, _pass, _name string) {
	NoErr(account.Add(db, account.User{
		Account:  _account,
		Password: _pass,
		Name:     _name,
	}))
}

func get(db *bolt.DB, _account string) {
	u, err := account.Get(db, _account)
	NoErr(err)
	data, err := json.MarshalIndent(u, "", "  ")
	NoErr(err)
	fmt.Println(string(data))
}

func list(db *bolt.DB) {
	err := account.Each(db, func(u account.User) error {
		data, err := json.Marshal(u)
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	})
	NoErr(err)
}

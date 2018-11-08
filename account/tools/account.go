package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
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
	defer db.Close()

	account.Prepare(db)
	list(db)
	add(db, "mofon", "zbiloveu", "admin")
	add(db, "cxy", "17743044405", "曹馨月")

	fmt.Println("account name:")
	for {
		var name, account string
		_, err := fmt.Scanf("%s %s", &account, &name)
		if err != nil {
			break
		}
		defaultPass := account
		add(db, account, defaultPass, name)
	}
}

func NoErr(err error) {
	return
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func add(db *bolt.DB, _account, _pass, _name string) {
	NoErr(db.Update(func(tx *bolt.Tx) error {
		return account.Add(tx, account.User{
			Account:  _account,
			Password: _pass,
			Name:     _name,
		})
	}))
}

func get(db *bolt.DB, _account string) {
	db.View(func(tx *bolt.Tx) error {
		u, err := account.Get(tx, _account)
		NoErr(err)
		data, err := json.MarshalIndent(u, "", "  ")
		NoErr(err)
		fmt.Println(string(data))
		return nil
	})
}

func list(db *bolt.DB) {
	db.View(func(tx *bolt.Tx) error {
		var users []account.User
		account.Each(tx, func(u account.User) error {
			users = append(users, u)
			return nil
		})

		sort.Slice(users, func(i, j int) bool {
			a, _ := strconv.Atoi(users[i].Id)
			b, _ := strconv.Atoi(users[j].Id)
			return a < b
		})

		for _, u := range users {
			fmt.Println(u.Id, u.Account, u.Name, u.Password)
		}

		return nil
	})
}

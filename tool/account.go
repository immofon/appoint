package main

import (
	"fmt"
	"os"
	"time"

	"github.com/immofon/appoint"
	"github.com/immofon/appoint/account"
	"github.com/immofon/appoint/log"
	bolt "go.etcd.io/bbolt"
)

func main() {
	log.TextMode()
	db, err := bolt.Open("appoint.bolt", 0600, &bolt.Options{
		Timeout: time.Second * 1,
	})
	NoErr(err)
	defer db.Close()

	account.Prepare(db)
	appoint.Prepare(db)
	add(db, "mofon", "zbiloveu", "admin")
	set_role(db, "mofon", appoint.Role_Admin)
	add(db, "cxy", "17743044405", "曹馨月")
	set_role(db, "cxy", appoint.Role_Teacher)

	fmt.Println("学号 姓名:")
	for {
		var name, account string
		_, err := fmt.Scanf("%s %s", &account, &name)
		if err != nil {
			break
		}
		defaultPass := generateDefaultPassword(account)
		fmt.Println(account, defaultPass)
		add(db, account, defaultPass, name)
		set_role(db, account, appoint.Role_Student)
	}
}

func generateDefaultPassword(account string) string {
	var a uint64
	_, err := fmt.Sscan(account, &a)
	if err != nil {
		panic(fmt.Sprint("error scanning value:", err))
	} else {
		return fmt.Sprint(a*11 - 97)
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
		id := account.Id(tx, _account)
		if id != "" {
			fmt.Println("account existed:", _account, _name)
		}
		return account.Add(tx, account.User{
			Account:  _account,
			Password: _pass,
			Name:     _name,
		})
	}))
}
func set_role(db *bolt.DB, _account string, role appoint.Role) {
	NoErr(db.Update(func(tx *bolt.Tx) error {
		return appoint.SetRole(tx, account.Id(tx, _account), role)
	}))
}

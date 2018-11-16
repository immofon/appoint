package appoint

import (
	"fmt"
	"testing"
	"time"

	"github.com/immofon/appoint/log"
	bolt "go.etcd.io/bbolt"
)

func Test_SetRole(t *testing.T) {
	log.TextMode()
	db, err := bolt.Open("tmp.bolt", 0600, &bolt.Options{
		Timeout: time.Second * 1,
	})
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	Prepare(db)
	db.Update(func(tx *bolt.Tx) error {
		SetRole(tx, "1", Role_Admin)
		SetRole(tx, "2", Role_Teacher)
		SetRole(tx, "3", Role_Teacher)
		SetRole(tx, "4", Role_Teacher)
		SetRole(tx, "5", Role_Teacher)
		SetRole(tx, "6", Role_Teacher)
		SetRole(tx, "7", Role_Teacher)
		fmt.Println(GetRole(tx, "6") == Role_Admin)
		EachRole(tx, func(id string, r Role) error {
			fmt.Println(id, ":", r)
			return nil
		})
		return nil
	})

}

func Test_IsCollided(t *testing.T) {
	log.TextMode()
	trs := make([]TimeRange, 0)
	for i := 1; i <= 10; i++ {
		tr := TimeRange{
			From: int64(i * 100),
			To:   int64(i*100 + 50),
		}
		trs = append(trs, tr)
	}
	fmt.Println("is", IsCollided(trs, TimeRange{
		From: 2,
		To:   3,
	}))

	fmt.Println("is", IsCollided(trs, TimeRange{
		From: 160,
		To:   160,
	}))
	fmt.Println("is", IsCollided(nil, TimeRange{
		From: 160,
		To:   1600,
	}))
}

func Test_Insert(t *testing.T) {
	log.TextMode()
	db, err := bolt.Open("tmp.bolt", 0x600, &bolt.Options{
		Timeout: time.Second * 1,
	})
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	Prepare(db)
	for i := 0; i < 100; i++ {
		db.Update(func(tx *bolt.Tx) error {
			return Insert(tx, TimeRange{
				From:    int64(i * 1000),
				To:      int64(i*1000 + 900),
				Teacher: "2",
				Student: "",
				Status:  Status_Enable,
			})
		})
	}
}

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
	db, err := bolt.Open("tmp.bolt", 0x600, &bolt.Options{
		Timeout: time.Second * 1,
	})
	if err != nil {
		t.Error(err)
	}
	defer db.Close()

	Prepare(db)
	SetRole(db, "1", Role_Admin)
	SetRole(db, "2", Role_Teacher)
	SetRole(db, "3", Role_Teacher)
	SetRole(db, "4", Role_Teacher)
	SetRole(db, "5", Role_Teacher)
	SetRole(db, "6", Role_Teacher)
	SetRole(db, "7", Role_Teacher)

	fmt.Println(GetRole(db, "6") == Role_Admin)

	EachRole(db, func(id string, r Role) error {
		fmt.Println(id, ":", r)
		return nil
	})
}

func Test_IsCollided(t *testing.T) {
	log.TextMode()
	trs := make([]TimeRange, 0)
	for i := 1; i <= 10; i++ {
		tr := TimeRange{
			From: uint64(i * 100),
			To:   uint64(i*100 + 50),
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
	Insert(db, TimeRange{
		From:    900,
		To:      2300,
		Teacher: "2",
		Student: "5",
	})
}

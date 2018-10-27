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

	Prepare(db)
	err = SetRole(db, "1", Role_Admin)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(GetRole(db, "12") == Role_Admin)
}

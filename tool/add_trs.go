package main

import (
	"time"

	"github.com/immofon/appoint"
	"github.com/immofon/appoint/log"
	"github.com/immofon/appoint/utils"
	"go.etcd.io/bbolt"
)

func main() {
	log.TextMode()
	db, _ := bbolt.Open("appoint.bolt", 0600, &bbolt.Options{
		Timeout: time.Second * 1,
	})
	defer db.Close()

	appoint.Prepare(db)
	trs := utils.NormalTimeRange(utils.NextDay(utils.DateZero(time.Now())))
	db.Update(func(tx *bbolt.Tx) error {
		for _, _tr := range trs {
			appoint.Insert(tx, appoint.TimeRange{
				From:    _tr.From,
				To:      _tr.To,
				Teacher: "2",
				Student: "",
				Status:  appoint.Status_Disable,
			})
		}
		return nil
	})
}

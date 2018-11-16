package utils

import (
	"fmt"
	"testing"
	"time"
)

func Test_Today(t *testing.T) {
	d := DateZero(time.Now())
	d = d.Add(time.Hour * 8)

	for i := 0; i < 6; i++ {
		from := d.Unix()
		to := d.Add(time.Minute*30).Unix() - 1

		fmt.Println(d)
		fmt.Println("from:", from, "to:", to)

		d = d.Add(time.Minute * 30)
	}
}

func Test_NormalTimeRange(t *testing.T) {
	trs := NormalTimeRange(time.Now())
	for _, tr := range trs {
		fmt.Println("from:", time.Unix(tr.From, 0), "-> to:", time.Unix(tr.To, 0))
	}
}

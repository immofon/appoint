package appoint

import (
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

func Benchmark_Generate(b *testing.B) {
	db, err := bolt.Open("tmp.bolt", 0x600, &bolt.Options{
		Timeout: time.Second * 1,
	})
	if err != nil {
		b.Error(err)
	}
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		UpdateData(tx)
		return nil
	})
	for i := 0; i < b.N; i++ {
		GetData()
	}
}

func h(data Data) {}

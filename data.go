package appoint

import (
	"sync"

	bolt "go.etcd.io/bbolt"
)

type UserState string

const (
	UserState_Unappointed = "unappointed"
	UserState_Appointed   = "appointed"
	UserState_Done        = "done"
)

type Data struct {
	UserStatus map[string]UserState
}

func newData() Data {
	return Data{
		UserStatus: make(map[string]UserState),
	}
}

var (
	data       Data          = newData()
	data_rwmux *sync.RWMutex = &sync.RWMutex{}
)

func UpdateData(tx *bolt.Tx) {
	data_rwmux.Lock()
	defer data_rwmux.Unlock()

	new_data := newData()

	EachRole(tx, func(id string, r Role) error {
		if r == Role_Student {
			new_data.UserStatus[id] = UserState_Unappointed
		}
		return nil
	})

	EachTimeRange(tx, func(tr TimeRange) error {
		sid := tr.Student
		if sid == "" {
			return nil
		}
		if tr.Status == Status_Achieved {
			new_data.UserStatus[sid] = UserState_Done
		} else if tr.Status == Status_Disable {
			new_data.UserStatus[sid] = UserState_Appointed
		}
		return nil
	})

	data = new_data
}

func GetData() Data {
	data_rwmux.RLock()
	defer data_rwmux.RUnlock()
	return data
}

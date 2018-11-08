package appoint

import (
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

func Generate(tx *bolt.Tx) Data {
	data := newData()

	EachRole(tx, func(id string, r Role) error {
		if r == Role_Student {
			data.UserStatus[id] = UserState_Unappointed
		}
		return nil
	})

	EachTimeRange(tx, func(tr TimeRange) error {
		sid := tr.Student
		if sid == "" {
			return nil
		}
		if tr.Status == Status_Achieved {
			data.UserStatus[sid] = UserState_Done
		} else if tr.Status == Status_Disable {
			data.UserStatus[sid] = UserState_Appointed
		}
		return nil
	})

	return data
}

package appoint

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/immofon/appoint/log"
	"github.com/immofon/appoint/utils"
	bolt "go.etcd.io/bbolt"
)

var (
	ErrTimeCollided = errors.New("time-collided")
)

const (
	bucket_appointment = "appointment"
	bucket_role        = "role"       // appointment/role
	bucket_time_range  = "time_range" // appointment/time_range
)

type Status string

const (
	Status_Enable   Status = "enable"
	Status_Disable  Status = "disable"
	Status_Achieved Status = "achieved"
	Status_Breached Status = "breached"
)

type Role string

const (
	Role_None    Role = ""
	Role_Admin   Role = "admin"
	Role_Teacher Role = "teacher"
	Role_Student Role = "student"
)

type TimeRange struct {
	From    int64  `json:"from"`    // unix
	To      int64  `json:"to"`      // unix
	Teacher string `json:"teacher"` // id
	Student string `json:"student"` // id
	Status  Status `json:"status"`
}

//Require: db.Update
//Error: utils.ErrInternal
func SetRole(tx *bolt.Tx, id string, r Role) error {
	b := tx.Bucket([]byte(bucket_appointment))
	if b == nil {
		log.L().Error()
		return utils.ErrInternal
	}

	b_role := b.Bucket([]byte(bucket_role))
	if b_role == nil {
		log.L().Error()
		return utils.ErrInternal
	}

	l := log.L().WithField("id", id).
		WithField("role", r)

	err := b_role.Put([]byte(id), []byte(r))
	if err != nil {
		l.Error(err)
		return utils.ErrInternal
	}

	l.Info("ok")
	return nil
}

//Error: utils.ErrInternal
func Prepare(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket_appointment))
		if err != nil {
			log.E(err).Error()
			return utils.ErrInternal
		}

		if _, err := b.CreateBucketIfNotExists([]byte(bucket_role)); err != nil {
			log.E(err).Error()
			return utils.ErrInternal
		}

		if _, err := b.CreateBucketIfNotExists([]byte(bucket_time_range)); err != nil {
			log.E(err).Error()
			return utils.ErrInternal
		}
		return nil
	})
}

func (tr TimeRange) Operable() bool {
	if tr.Status == Status_Enable {
		return true
	}
	if tr.Status == Status_Disable && tr.Student == "" {
		return true
	}
	return false
}

//Require: db.View
func GetRole(tx *bolt.Tx, id string) Role {
	b := tx.Bucket([]byte(bucket_appointment))
	if b == nil {
		log.L().Error()
		return Role_None
	}

	b_role := b.Bucket([]byte(bucket_role))
	if b_role == nil {
		log.L().Error()
		return Role_None
	}

	return Role(b_role.Get([]byte(id)))
}

//Require: db.View
//Error: inside-fn|utils.ErrInternal
func EachRole(tx *bolt.Tx, fn func(id string, r Role) error) error {
	b := tx.Bucket([]byte(bucket_appointment))
	if b == nil {
		log.L().Error()
		return utils.ErrInternal
	}

	b_role := b.Bucket([]byte(bucket_role))
	if b_role == nil {
		log.L().Error()
		return utils.ErrInternal
	}

	return b_role.ForEach(func(_id, _role []byte) error {
		return fn(string(_id), Role(_role))
	})
}

//Require: db.Update
func Insert(tx *bolt.Tx, tr TimeRange) error {
	b := tx.Bucket([]byte(bucket_appointment))
	if b == nil {
		log.L().Error()
		return utils.ErrInternal
	}

	b_tr := b.Bucket([]byte(bucket_time_range))
	if b_tr == nil {
		log.L().Error()
		return utils.ErrInternal
	}

	trs := make([]TimeRange, 0, 100)
	b_tr.ForEach(func(_, _tr []byte) error {
		var tr TimeRange
		json.Unmarshal(_tr, &tr)
		trs = append(trs, tr)
		return nil
	})

	if IsCollided(trs, tr) {
		log.E(ErrTimeCollided).Error()
		return ErrTimeCollided
	}

	// just insert
	data, err := json.Marshal(tr)
	if err != nil {
		log.E(err).Error()
		return utils.ErrInternal
	}
	if err := b_tr.Put([]byte(TimeRangeId(tr)), data); err != nil {
		log.E(err).Error()
		return utils.ErrInternal
	}
	log.L().WithField("from", tr.From).
		WithField("to", tr.To).
		WithField("teacher", tr.Teacher).
		WithField("student", tr.Student).
		WithField("status", tr.Status).
		Info("ok")
	return nil
}

func UpdateTimeRange(tx *bolt.Tx, tr_id string, fn func(TimeRange) (TimeRange, error)) error {
	b := tx.Bucket([]byte(bucket_appointment))
	if b == nil {
		log.L().Error()
		return utils.ErrInternal
	}

	b_tr := b.Bucket([]byte(bucket_time_range))
	if b_tr == nil {
		log.L().Error()
		return utils.ErrInternal
	}

	// read
	var tr TimeRange
	if err := json.Unmarshal(b_tr.Get([]byte(tr_id)), &tr); err != nil {
		log.E(err).Error()
		return utils.ErrInternal
	}

	// handle
	tr, err := fn(tr)
	if err != nil {
		return err
	}

	// update
	data, err := json.Marshal(tr)
	if err != nil {
		log.E(err).Error()
		return utils.ErrInternal
	}
	if err := b_tr.Put([]byte(TimeRangeId(tr)), data); err != nil {
		log.E(err).Error()
		return utils.ErrInternal
	}

	return nil
}

// require: db.View
// Error: err in fn | utils.ErrInternal
func EachTimeRange(tx *bolt.Tx, fn func(TimeRange) error) error {
	b := tx.Bucket([]byte(bucket_appointment))
	if b == nil {
		log.L().Error()
		return utils.ErrInternal
	}

	b_time_range := b.Bucket([]byte(bucket_time_range))
	if b_time_range == nil {
		log.L().Error()
		return utils.ErrInternal
	}

	return b_time_range.ForEach(func(_id, _tr []byte) error {
		var tr TimeRange
		if err := json.Unmarshal(_tr, &tr); err != nil {
			log.E(err).
				WithField("data", _tr).
				WithField("data.string", string(_tr)).
				Error("json.Unmarshal failure")
			return utils.ErrInternal
		}
		return fn(tr)
	})
}

func IsCollided(trs []TimeRange, tr TimeRange) bool {
	if tr.From >= tr.To {
		return true
	}

	trs = append(trs, tr)

	sort.Slice(trs, func(i, j int) bool {
		return trs[i].From < trs[j].From
	})

	haveNext := func(i int) bool {
		return (i + 1) < len(trs)
	}

	for i, tr := range trs {
		if !haveNext(i) {
			continue
		}

		next := trs[i+1]

		if tr.From <= next.From && next.From <= tr.To {
			return true
		}
	}
	return false
}

func TimeRangeId(tr TimeRange) string {
	return fmt.Sprintf("%d_%d", tr.From, tr.To)
}

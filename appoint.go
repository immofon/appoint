package appoint

import (
	"github.com/immofon/appoint/log"
	"github.com/immofon/appoint/utils"
	bolt "go.etcd.io/bbolt"
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

type TimmRange struct {
	From    uint64 `json:"from"`    // unix
	To      uint64 `json:"to"`      // unix
	Teacher string `json:"teacher"` // id
	Student string `json:"student"` // id
	Status  Status `json:"status"`
}

func SetRole(db *bolt.DB, id string, r Role) error {
	return db.Update(func(tx *bolt.Tx) error {
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
	})
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

func GetRole(db *bolt.DB, id string) Role {
	r := Role_None
	db.View(func(tx *bolt.Tx) error {
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

		r = Role(b_role.Get([]byte(id)))
		return nil
	})
	return r
}

package account

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strconv"

	"errors"

	"github.com/immofon/appoint/log"
	bolt "go.etcd.io/bbolt"
)

// No field could be empty
type User struct {
	Id string `json:"id"` // unique;auto-generate

	Account  string `json:"account"`  // unique;no-surround-space
	Password string `json:"password"` // md5

	Name string `json:"name"` //no-surround-space
}

var (
	ErrInternal = errors.New("internal-error")
	ErrNotFound = errors.New("not-found")
	ErrUnvalid  = errors.New("unvalid")

	ErrNotSet       = errors.New("not-set")
	ErrAccountExist = errors.New("account-exist")

	ErrNotFoundAccount = errors.New("not-found-account")
	ErrNotFoundBucket  = errors.New("not-found-bucket")
)

var (
	bucket_account    = "account"
	bucket_account2id = "__$account2id" // nested in bucket_account
	key_next_id       = "__internal_next_id"
	default_next_id   = "1"
)

//Error: ErrInternal|ErrNotSet|ErrAccountExist
func Add(db *bolt.DB, u User) error {
	if u.Account == "" || u.Password == "" || u.Name == "" {
		return ErrNotSet
	}
	// md5 password
	u.Password = md5pass(u.Account, u.Password)

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket_account))
		if b == nil {
			log.L().
				WithField("bucket", bucket_account).
				Error("no bucket")
			return ErrInternal
		}

		b_a2id := b.Bucket([]byte(bucket_account2id))
		if b_a2id == nil {
			log.L().
				WithField("bucket", bucket_account2id).
				Error("no bucket")
			return ErrInternal
		}

		// make sure account don't exists
		if b_a2id.Get([]byte(u.Account)) != nil {
			log.L().
				WithField("account", u.Account).
				Info("account exists")
			return ErrAccountExist
		}

		// generate id
		next_id, err := strconv.Atoi(string(b.Get([]byte(key_next_id))))
		if err != nil {
			log.E(err).Error("content of", key_next_id, "was not int")
			return ErrInternal
		}

		u.Id = strconv.Itoa(next_id)

		// set next_id
		err = b.Put([]byte(key_next_id), []byte(strconv.Itoa(next_id+1)))
		if err != nil {
			log.E(err).Error()
			return ErrInternal
		}

		data, err := json.Marshal(u)
		if err != nil {
			log.E(err).Error()
			return ErrInternal
		}

		//account: [id]=>[user;json]
		if err = b.Put([]byte(u.Id), data); err != nil {
			log.E(err).Error()
			return ErrInternal
		}

		//account.account2id: [account]=>[id]
		if err := b_a2id.Put([]byte(u.Account), []byte(u.Id)); err != nil {
			log.E(err).Error()
			return ErrInternal
		}

		log.L().WithField("id", u.Id).
			WithField("account", u.Account).
			WithField("name", u.Name).Info()
		return nil
	})
}

func Get(db *bolt.DB, id string) (User, error) {
	var u User
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket_account))
		if b == nil {
			log.L().
				WithField("bucket", bucket_account).
				Error("no bucket")
			return ErrInternal
		}

		data := b.Get([]byte(id))
		if data == nil {
			log.L().
				WithField("id", id).
				Debug("not found id in account bucket")
			return ErrNotFound
		}
		return json.Unmarshal(data, &u)
	})
	if err == nil {
		log.L().
			WithField("id", id).
			WithField("account", u.Account).
			WithField("name", u.Name).
			Debug("ok")
		return u, nil
	} else {
		log.E(err).Error()
		return u, ErrInternal
	}
}

func Each(db *bolt.DB, fn func(User) error) error {
	return db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket_account))
		if b == nil {
			log.L().
				WithField("bucket", bucket_account).
				Error("no bucket")
			return ErrInternal
		}

		return b.ForEach(func(key, _user_json []byte) error {
			if is_internal_key(key) {
				return nil
			}
			var u User
			err := json.Unmarshal(_user_json, &u)
			if err != nil {
				return err
			}
			return fn(u)
		})
	})
}

//Error: ErrInternal|ErrNotFound
func Id(db *bolt.DB, account string) (id string) {
	if account == "" {
		return ""
	}

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket_account))
		if b == nil {
			log.L().
				WithField("bucket", bucket_account).
				Error("no bucket")
			return ErrInternal
		}

		b_a2id := b.Bucket([]byte(bucket_account2id))
		if b_a2id == nil {
			log.L().
				WithField("bucket", bucket_account2id).
				Error("no bucket")
			return ErrInternal
		}

		id = string(b_a2id.Get([]byte(account)))
		if id == "" {
			return ErrNotFound
		}
		return nil
	})
	return id
}

func Prepare(db *bolt.DB) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket_account))
		if err != nil {
			return ErrInternal
		}

		if b.Get([]byte(key_next_id)) == nil {
			if err := b.Put([]byte(key_next_id), []byte(default_next_id)); err != nil {
				return ErrInternal
			}
		}

		if _, err := b.CreateBucketIfNotExists([]byte(bucket_account2id)); err != nil {
			return ErrInternal
		}
		return nil
	})
}

//Error: .Get|ErrUnvalid
func Auth(db *bolt.DB, account, password string) (u User, err error) {
	u, err = Get(db, Id(db, account))
	if err != nil {
		return u, err
	}

	if u.Password == md5pass(account, password) {
		log.L().
			WithField("id", u.Id).
			WithField("account", u.Account).
			Info("ok")
		return u, nil
	} else {
		log.E(ErrUnvalid).Debug()
		return u, ErrUnvalid
	}
}

func md5pass(account, password string) string {
	const Magic = "$#*()*$"
	data := md5.Sum([]byte(Magic + account + password + account + Magic))
	return fmt.Sprintf("%x", data)
}

func is_internal_key(k []byte) bool {
	return bytes.HasPrefix(k, []byte("__"))

}

package main

import (
	"time"

	"github.com/immofon/appoint"
)

func isTeacherSchedule(tr appoint.TimeRange, teacher string, after time.Time) bool {
	after_stamp := after.Unix()
	if tr.Status != appoint.Status_Disable {
		return false
	}

	if tr.Student == "" || tr.Teacher != teacher {
		return false
	}

	if tr.To < after_stamp {
		return false
	}
	return true
}

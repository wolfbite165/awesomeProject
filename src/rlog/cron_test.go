package rlog

import (
	"log"
	"testing"
	"time"
)

func TestMatchTime6(t *testing.T) {
	Help_SetLogOutput(t)
	jobs := NewJobList()
	jobs.AddJob("0 0 1 3 *", nil)
	job := jobs.jobSlice[0]
	for i := time.Now(); i.Before(time.Now().AddDate(0, 10, 20)); i = i.Add(time.Second * 10) {
		if job.IsMatchTime(i) {
			log.Printf("i=%v,dow=%v\n", i, i.Weekday())
		}
	}
}

func TestMatchTime5(t *testing.T) {
	Help_SetLogOutput(t)
	jobs := NewJobList()
	jobs.AddJob("0 0 15 * *", nil)
	job := jobs.jobSlice[0]
	for i := time.Now(); i.Before(time.Now().AddDate(0, 10, 20)); i = i.Add(time.Second * 10) {
		if job.IsMatchTime(i) {
			log.Printf("i=%v,dow=%v\n", i, i.Weekday())
		}
	}
}

func TestMatchTime4(t *testing.T) {
	Help_SetLogOutput(t)
	jobs := NewJobList()
	jobs.AddJob("0 0 * * 1", nil)
	job := jobs.jobSlice[0]
	for i := time.Now(); i.Before(time.Now().AddDate(0, 0, 20)); i = i.Add(time.Second) {
		if job.IsMatchTime(i) {
			log.Printf("i=%v,dow=%v\n", i, i.Weekday())
		}
	}
}

func TestMatchTime3(t *testing.T) {
	Help_SetLogOutput(t)
	jobs := NewJobList()
	jobs.AddJob("0 0 * * *", nil)
	job := jobs.jobSlice[0]
	for i := time.Now(); i.Before(time.Now().AddDate(0, 0, 10)); i = i.Add(time.Second) {
		if job.IsMatchTime(i) {
			log.Println(i)
		}
	}
}

func TestMatchTime2(t *testing.T) {
	Help_SetLogOutput(t)
	jobs := NewJobList()
	jobs.AddJob("*/3 */4 * * *", nil)
	job := jobs.jobSlice[0]
	for i := time.Now(); i.Before(time.Now().AddDate(0, 0, 1)); i = i.Add(time.Second) {
		if job.IsMatchTime(i) {
			log.Println(i)
		}
	}
}

func TestMatchTime1(t *testing.T) {
	Help_SetLogOutput(t)
	jobs := NewJobList()
	jobs.AddJob("50 * * * *", nil)
	job := jobs.jobSlice[0]
	for i := time.Now(); i.Before(time.Now().AddDate(0, 0, 1)); i = i.Add(time.Second * 10) {
		if job.IsMatchTime(i) {
			if i.Minute() != 50 {
				log.Panic("匹配出错")
			}
		}
	}
}

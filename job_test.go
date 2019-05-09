// Copyright 2019 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package scheduled

import (
	"errors"
	"io/ioutil"
	"log"
	"testing"
	"time"

	"github.com/issue9/assert"

	"github.com/issue9/scheduled/schedulers/ticker"
)

var (
	succFunc = func() error {
		println("succ", time.Now().String())
		return nil
	}

	erroFunc = func() error {
		println("erro", time.Now().String())
		return errors.New("erro")
	}

	failFunc = func() error {
		println("fail", time.Now().String())
		panic("fail")
	}

	errlog = log.New(ioutil.Discard, "ERRO", 0)
)

func TestJob_run(t *testing.T) {
	a := assert.New(t)
	now := time.Now()

	j := &Job{
		name:      "succ",
		f:         succFunc,
		scheduler: ticker.New(time.Second),
	}
	j.init(now)
	j.run(now, nil)
	a.Nil(j.Err()).
		Equal(j.State(), Stoped).
		True(j.next.After(now)).
		True(j.next.After(j.prev))

	j = &Job{
		name:      "erro",
		f:         erroFunc,
		scheduler: ticker.New(time.Second),
	}
	j.init(now)
	j.run(now, errlog)
	a.NotNil(j.Err()).
		Equal(j.State(), Failed).
		True(j.next.After(now)).
		True(j.next.After(j.prev))

	j = &Job{
		name:      "fail",
		f:         failFunc,
		scheduler: ticker.New(time.Second),
	}
	j.init(now)
	j.run(now, nil)
	a.NotNil(j.Err()).
		Equal(j.State(), Failed).
		True(j.next.After(now)).
		True(j.next.After(j.prev))
}

func TestSortJobs(t *testing.T) {
	a := assert.New(t)

	now := time.Now()
	jobs := []*Job{
		&Job{
			name: "1",
			next: now.Add(1111),
		},
		&Job{
			name: "2",
			next: time.Time{},
		},
		&Job{
			name: "3",
			next: now,
		},
		&Job{
			name: "4",
			next: time.Time{},
		},
		&Job{
			name: "5",
			next: now.Add(222),
		},
	}

	sortJobs(jobs)
	a.Equal(jobs[0].name, "3").
		Equal(jobs[1].name, "5").
		Equal(jobs[2].name, "1")
}

func TestServer_NewCron(t *testing.T) {
	a := assert.New(t)

	srv := NewServer(nil)
	a.NotError(srv.NewCron("test", nil, "* * * 3-7 * *"))
	a.Error(srv.NewCron("test", nil, "* * * 3-7a * *"))
}

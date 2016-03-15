package main

import (
	"github.com/benbjohnson/clock"
	"reflect"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	mock := clock.NewMock()
	mock.Add(10 * time.Hour) // clock is now 1970-01-01 10:00:00 +0000 UTC

	cases := []struct {
		ch  string
		usr string
		msg string
		m   *Memo
		err error
	}{
		// test empty cases
		{
			"chanName",
			"userName",
			"",
			nil,
			errEmpty,
		},
		{
			"chanName",
			"userName",
			"  ",
			nil,
			errEmpty,
		},
		{
			"chanName",
			"userName",
			"	",
			nil,
			errEmpty,
		},
		// standard case
		{
			"chanName",
			"userName",
			"message",
			&Memo{
				Date: time.Unix(10*60*60-25, 0).UTC(),
				Desc: "message",
				Tags: []string{"author:userName", "chan:chanName"},
			},
			nil,
		},
		// standard with extraneous whitespace
		{
			"chanName",
			"userName",
			"   some message ",
			&Memo{
				Date: time.Unix(10*60*60-25, 0).UTC(),
				Desc: "some message",
				Tags: []string{"author:userName", "chan:chanName"},
			},
			nil,
		},
		// same but empty chan
		{
			"null",
			"userName",
			"   some message ",
			&Memo{
				Date: time.Unix(10*60*60-25, 0).UTC(),
				Desc: "some message",
				Tags: []string{"author:userName"},
			},
			nil,
		},
		// override default offset
		{
			"chanName",
			"userName",
			" 0 some message",
			&Memo{
				Date: time.Unix(10*60*60, 0).UTC(),
				Desc: "some message",
				Tags: []string{"author:userName", "chan:chanName"},
			},
			nil,
		},
		// custom offset
		{
			"chanName",
			"userName",
			" 1 some message",
			&Memo{
				Date: time.Unix(10*60*60-1, 0).UTC(),
				Desc: "some message",
				Tags: []string{"author:userName", "chan:chanName"},
			},
			nil,
		},
		// more interesting timespec
		{
			"chanName",
			"userName",
			" 5min3s some message",
			&Memo{
				Date: time.Unix(10*60*60-5*60-3, 0).UTC(),
				Desc: "some message",
				Tags: []string{"author:userName", "chan:chanName"},
			},
			nil,
		},
		// same, but combined with extra tag
		{
			"chanName",
			"userName",
			" 5min3s some message some:tag",
			&Memo{
				Date: time.Unix(10*60*60-5*60-3, 0).UTC(),
				Desc: "some message",
				Tags: []string{"author:userName", "chan:chanName", "some:tag"},
			},
			nil,
		},
		// full date-time spec and extra tag
		{
			"chanName",
			"userName",
			" 1970-01-01T12:34:56Z some message some:tag xyz:tag",
			&Memo{
				Date: time.Unix(12*3600+34*60+56, 0).UTC(),
				Desc: "some message",
				Tags: []string{"author:userName", "chan:chanName", "some:tag", "xyz:tag"},
			},
			nil,
		},
		// same but try to override author tag
		{
			"chanName",
			"userName",
			" 1970-01-01T12:34:56Z some message some:tag author:someone-else",
			nil,
			errFixedTag,
		},
	}
	for i, c := range cases {
		out, err := getMemo(c.ch, c.usr, c.msg, mock)
		if !reflect.DeepEqual(c.err, err) {
			t.Errorf("case %d: bad err output\ninput: %#v\nexp %v\ngot %v", i, c.msg, c.err, err)
		}
		if !reflect.DeepEqual(c.m, out) {
			t.Errorf("case %d: bad memo output\ninput: %#v\nexp %v\ngot %v", i, c.msg, c.m, out)
		}
	}
}

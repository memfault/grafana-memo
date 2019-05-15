package daemon

import (
	"reflect"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/raintank/memo"
)

func TestParse(t *testing.T) {
	mock := clock.NewMock()
	mock.Add(10 * time.Hour) // clock is now 1970-01-01 10:00:00 +0000 UTC

	cases := []struct {
		ch     string
		usr    string
		msg    string
		expM   memo.Memo
		expErr error
	}{
		// test empty cases
		{
			"chanName",
			"userName",
			"",
			memo.Memo{},
			errEmpty,
		},
		{
			"chanName",
			"userName",
			"  ",
			memo.Memo{},
			errEmpty,
		},
		{
			"chanName",
			"userName",
			"	",
			memo.Memo{},
			errEmpty,
		},
		// standard case
		{
			"chanName",
			"userName",
			"message",
			memo.Memo{
				Date: time.Unix(10*60*60-25, 0).UTC(),
				Desc: "message",
				Tags: []string{"memo", "author:userName", "chan:chanName"},
			},
			nil,
		},
		// standard with extraneous whitespace
		{
			"chanName",
			"userName",
			"   some message ",
			memo.Memo{
				Date: time.Unix(10*60*60-25, 0).UTC(),
				Desc: "some message",
				Tags: []string{"memo", "author:userName", "chan:chanName"},
			},
			nil,
		},
		// same but empty chan
		{
			"null",
			"userName",
			"   some message ",
			memo.Memo{
				Date: time.Unix(10*60*60-25, 0).UTC(),
				Desc: "some message",
				Tags: []string{"memo", "author:userName"},
			},
			nil,
		},
		// override default offset
		{
			"chanName",
			"userName",
			" 0 some message",
			memo.Memo{
				Date: time.Unix(10*60*60, 0).UTC(),
				Desc: "some message",
				Tags: []string{"memo", "author:userName", "chan:chanName"},
			},
			nil,
		},
		// custom offset
		{
			"chanName",
			"userName",
			" 1 some message",
			memo.Memo{
				Date: time.Unix(10*60*60-1, 0).UTC(),
				Desc: "some message",
				Tags: []string{"memo", "author:userName", "chan:chanName"},
			},
			nil,
		},
		// more interesting timespec
		{
			"chanName",
			"userName",
			" 5min3s some message",
			memo.Memo{
				Date: time.Unix(10*60*60-5*60-3, 0).UTC(),
				Desc: "some message",
				Tags: []string{"memo", "author:userName", "chan:chanName"},
			},
			nil,
		},
		// same, but combined with extra tag
		{
			"chanName",
			"userName",
			" 5min3s some message some:tag",
			memo.Memo{
				Date: time.Unix(10*60*60-5*60-3, 0).UTC(),
				Desc: "some message",
				Tags: []string{"memo", "author:userName", "chan:chanName", "some:tag"},
			},
			nil,
		},
		// full date-time spec and extra tag
		{
			"chanName",
			"userName",
			" 1970-01-01T12:34:56Z some message some:tag xyz:tag",
			memo.Memo{
				Date: time.Unix(12*3600+34*60+56, 0).UTC(),
				Desc: "some message",
				Tags: []string{"memo", "author:userName", "chan:chanName", "some:tag", "xyz:tag"},
			},
			nil,
		},
		// same but try to override author tag
		{
			"chanName",
			"userName",
			" 1970-01-01T12:34:56Z some message some:tag author:someone-else",
			memo.Memo{},
			errFixedTag,
		},
	}
	for i, c := range cases {
		out, err := ParseMemo(c.ch, c.usr, c.msg, mock)
		if !reflect.DeepEqual(c.expErr, err) {
			t.Errorf("case %d: bad err output\ninput: %#v\nexp %v\ngot %v", i, c.msg, c.expErr, err)
		}
		if !reflect.DeepEqual(c.expM, out) {
			t.Errorf("case %d: bad memo output\ninput: %#v\nexp %v\ngot %v", i, c.msg, c.expM, out)
		}
	}
}
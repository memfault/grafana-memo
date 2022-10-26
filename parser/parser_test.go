package parser

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/grafana/memo"
)

func TestParse(t *testing.T) {
	mock := clock.NewMock()
	mock.Add(10 * time.Hour) // clock is now 1970-01-01 10:00:00 +0000 UTC

	cases := []struct {
		msg     string
		expErr  error
		expDate time.Time
		expDesc string
		expTags []string
	}{
		// test empty cases
		{
			msg:    "",
			expErr: memo.ErrEmpty,
		},
		{
			msg:    "  ",
			expErr: memo.ErrEmpty,
		},
		{
			msg:    "	",
			expErr: memo.ErrEmpty,
		},
		// standard case
		{
			msg:     "message",
			expDate: time.Unix(10*60*60-25, 0).UTC(),
			expDesc: "message",
			expTags: []string{},
		},
		// standard with extraneous whitespace
		{
			msg:     "   some message ",
			expDate: time.Unix(10*60*60-25, 0).UTC(),
			expDesc: "some message",
			expTags: []string{},
		},
		// override default offset
		{
			msg:     " 0 some message",
			expDate: time.Unix(10*60*60, 0).UTC(),
			expDesc: "some message",
			expTags: []string{},
		},
		// custom offset
		{
			msg:     " 1 some message",
			expDate: time.Unix(10*60*60-1, 0).UTC(),
			expDesc: "some message",
			expTags: []string{},
		},
		// more interesting timespec
		{
			msg:     " 5min3s some message",
			expDate: time.Unix(10*60*60-5*60-3, 0).UTC(),
			expDesc: "some message",
			expTags: []string{},
		},
		// same, but combined with extra tag
		{
			msg:     " 5min3s some message some:tag",
			expDate: time.Unix(10*60*60-5*60-3, 0).UTC(),
			expDesc: "some message",
			expTags: []string{"some:tag"},
		},
		// full date-time spec and extra tag
		{
			msg:     " 1970-01-01T12:34:56Z some message some:tag xyz:tag",
			expDate: time.Unix(12*3600+34*60+56, 0).UTC(),
			expDesc: "some message",
			expTags: []string{"some:tag", "xyz:tag"},
		},
	}

	parser := New()

	for i, c := range cases {
		memo, err := parser.Parse(c.msg)
		if !errors.Is(err, c.expErr) {
			t.Errorf("case %d: bad err output\ninput: %#v\nexp %v\ngot %v", i, c.msg, c.expErr, err)
		}

		if err != nil {
			continue
		}

		if memo.Date != c.expDate || memo.Desc != c.expDesc || !reflect.DeepEqual(c.expTags, memo.Tags) {
			fmt.Println(memo.Date != c.expDate)
			fmt.Println(memo.Desc != c.expDesc, !reflect.DeepEqual(c.expTags, memo.Tags))
			t.Errorf("case %d: bad output\ninput: %#v\nexp date=%s, desc=%q, tags=%v\ngot date=%s, desc=%q, tags=%v\n", i, c.msg, c.expDate, c.expDesc, c.expTags, memo.Date, memo.Desc, memo.Tags)
		}
	}
}

package parser

import (
	"errors"
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
		expNil  bool
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
		// test not for us cases
		{
			msg:    "this is just a message in chat",
			expNil: true,
		},
		// standard case
		{
			msg:     "memo message",
			expDate: time.Unix(10*60*60-25, 0),
			expDesc: "message",
			expTags: []string{"memo"},
		},
		// standard with extraneous whitespace
		{
			msg:     "  memo    some message ",
			expDate: time.Unix(10*60*60-25, 0),
			expDesc: "some message",
			expTags: []string{"memo"},
		},
		// override default offset
		{
			msg:     "memo 0 some message",
			expDate: time.Unix(10*60*60, 0),
			expDesc: "some message",
			expTags: []string{"memo"},
		},
		// custom offset
		{
			msg:     "memo 1 some message",
			expDate: time.Unix(10*60*60-1, 0),
			expDesc: "some message",
			expTags: []string{"memo"},
		},
		// more interesting timespec
		{
			msg:     "memo 5min3s some message",
			expDate: time.Unix(10*60*60-5*60-3, 0),
			expDesc: "some message",
			expTags: []string{"memo"},
		},
		// same, but combined with extra tag
		{
			msg:     "memo 5min3s some message some:tag",
			expDate: time.Unix(10*60*60-5*60-3, 0),
			expDesc: "some message",
			expTags: []string{"memo", "some:tag"},
		},
		// full date-time spec and extra tag
		{
			msg:     "memo 1970-01-01T12:34:56Z some message some:tag xyz:tag",
			expDate: time.Unix(12*3600+34*60+56, 0).UTC(),
			expDesc: "some message",
			expTags: []string{"memo", "some:tag", "xyz:tag"},
		},
	}

	parser := New()
	parser.SetClock(mock)

	for i, c := range cases {
		m, err := parser.Parse(c.msg)
		if !errors.Is(err, c.expErr) {
			t.Errorf("case %d: bad err output\ninput: %#v\nexp %v\ngot %v", i, c.msg, c.expErr, err)
		}

		if err != nil {
			continue
		}

		if m == nil {
			if !c.expNil {
				t.Errorf("we are expecting a memo, but received none")
			}

			continue
		}

		m.Date = m.Date.Round(time.Second)

		if m.Date != c.expDate || m.Desc != c.expDesc || !reflect.DeepEqual(c.expTags, m.Tags) {
			t.Errorf("case %d: bad output\ninput: %#v\nexp date=%s, desc=%q, tags=%v\ngot date=%s, desc=%q, tags=%v\n", i, c.msg, c.expDate, c.expDesc, c.expTags, m.Date, m.Desc, m.Tags)
		}
	}
}

package ini_test

import (
	"strconv"
	"testing"

	"github.com/gookit/goutil/x/assert"
	"github.com/gookit/ini/v2"
)

func TestUintFullRange(t *testing.T) {
	max := ^uint(0)
	for _, tc := range []struct {
		name, input string
		want        uint
		wantError   bool
	}{
		{"zero", "0", 0, false},
		{"ordinary value", "28", 28, false},
		{"maximum uint", strconv.FormatUint(uint64(max), 10), max, false},
		{"negative value", "-1", 0, true},
		{"overflow", strconv.FormatUint(uint64(max), 10) + "0", 0, true},
		{"invalid", "invalid", 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conf := ini.New()
			assert.NoErr(t, conf.Set("value", tc.input))
			assert.Eq(t, tc.want, conf.Uint("value", 7))
			assert.Eq(t, tc.wantError, conf.Error() != nil)
		})
	}
	t.Run("missing value", func(t *testing.T) {
		conf := ini.New()
		assert.Eq(t, uint(0), conf.Uint("missing"))
		assert.Eq(t, uint(7), conf.Uint("missing", 7))
		assert.NoErr(t, conf.Error())
	})
	if strconv.IntSize == 64 {
		t.Run("above signed maximum", func(t *testing.T) {
			conf := ini.New()
			assert.NoErr(t, conf.Set("value", "9223372036854775808"))
			got := conf.Uint("value")
			assert.Eq(t, uint64(9223372036854775808), uint64(got))
			assert.NoErr(t, conf.Error())
		})
	}
}

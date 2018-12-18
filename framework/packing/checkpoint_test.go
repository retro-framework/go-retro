package packing

import (
	"testing"
	"time"

	"github.com/retro-framework/go-retro/framework/test_helper"
)

// Note that an important part of these tests is that
// we only see families of error messages which make
// sense, it's no use seeing format and TZ errors for
// a date string that is empty, so we assert against
// an exact set.
func Test_Checkpoint_HasErrors(t *testing.T) {

	t.Run("date", func(t *testing.T) {

		var fixedDate, _ = time.Parse(
			time.RFC3339,
			"2012-11-01T22:08:41+00:00")

		t.Run("absent", func(t *testing.T) {
			t.Parallel()
			var h = test_helper.H(t)
			var hasErrs, errs = Checkpoint{}.HasErrors()
			t.Logf(h.MustSerilizeYAML(hasErrs))
			t.Logf("Errors: %s\n", errs)
		})

		t.Run("empty string", func(t *testing.T) {
			t.Parallel()
			var h = test_helper.H(t)
			var hasErrs, errs = Checkpoint{
				Fields: map[string]string{"date": ""},
			}.HasErrors()
			t.Logf(h.MustSerilizeYAML(hasErrs))
			t.Logf("Errors: %s\n", errs)
		})

		t.Run("not rfc3339", func(t *testing.T) {
			t.Parallel()
			var h = test_helper.H(t)
			var hasErrs, errs = Checkpoint{
				Fields: map[string]string{"date": fixedDate.Format(time.RFC822)},
			}.HasErrors()
			h.BoolEql(hasErrs, true)
			t.Logf(h.MustSerilizeYAML(hasErrs))
			t.Logf("Errors: %s\n", errs)
		})

		t.Run("rfc3339 but not UTC", func(t *testing.T) {
			t.Parallel()

			// China doesn't have daylight saving. It uses a fixed 8 hour offset from UTC.
			beijing := time.FixedZone("Beijing Time", int((8 * time.Hour).Seconds()))

			var h = test_helper.H(t)
			var hasErrs, errs = Checkpoint{
				Fields: map[string]string{"date": fixedDate.In(beijing).Format(time.RFC3339)},
			}.HasErrors()
			t.Logf(h.MustSerilizeYAML(hasErrs))
			t.Logf("Errors: %s\n", errs)
		})
	})

}

package packing

import (
	"time"

	"github.com/pkg/errors"
	"github.com/retro-framework/go-retro/framework/retro"
)

var (
	ErrCheckpointDateFieldEmptyString = errors.New("checkpoint cannot have an empty string for the value of the date field")
	ErrCheckpointDateMustParseRFC3999 = errors.New("checkpoint date field must be parseable as a RFC3339 date")
	ErrCheckpointTZMustBeUTC          = errors.New("checkpoint date field must in timezone UTC")
	ErrCheckpointTZOffsetMustBeZero   = errors.New("checkpoint date field not have a non-zero TZ offset (must be UTC)")
	ErrCheckpointDateFieldAbsent      = errors.New("checkpoint has no `date' field, cannot be saved")
)

// Checkpoint represents a DDD command object execution
// and persistence of the resulting events. It stores
// an error incase the command failed.
type Checkpoint struct {
	AffixHash    retro.Hash        `json:"affixHash"`
	ParentHashes []retro.Hash      `json:"parentHashes"`
	Fields       map[string]string `json:"fields"`
	Summary      string            `json:"summary"`
	CommandDesc  []byte            `json:"commandDesc"`
}

// HasErrors is used for example to determine if a Checkpoint
// has any errors which would preclude storing it. For now
// the only case is that it absolutely must have a Date and
// Session (case insensitive) field in the Fields.
func (c Checkpoint) HasErrors() (bool, []error) {
	var errs []error

	// the ["date"] entry on the fields is mandatory, it must be set
	// to something parseable as an RFC3339 date string. Timezone
	// must be set to UTC else an error will be raised. (Retro makes
	// no provision for storing the original timestamp and will
	// force everything to UTC internally, destroying data. This
	// error for timezones other than UTC prevents that)
	if v, exists := c.Fields["date"]; exists {
		if len(v) == 0 {
			errs = append(errs, ErrCheckpointDateFieldEmptyString)
			return len(errs) > 0, errs
		}
		if t, parserErr := time.Parse(time.RFC3339, v); parserErr != nil {
			// TODO: Would be good to find a way to pass the error
			// details back up to the caller other than a generic "parser
			// error"
			errs = append(errs, ErrCheckpointDateMustParseRFC3999)
		} else {
			if parserErr == nil {
				tzName, tzOffset := t.Zone()
				if tzName != "UTC" {
					errs = append(errs, ErrCheckpointTZMustBeUTC)
				} else if tzOffset > 0 {
					errs = append(errs, ErrCheckpointTZOffsetMustBeZero)
				}
			}
		}
	} else {
		errs = append(errs, ErrCheckpointDateFieldAbsent)
	}
	// TODO: Also ensure that Session is set..

	// TODO: check if this checkpoint has an affix, and use ErrCheckpointWithoutAffix if not.

	return len(errs) > 0, errs
}

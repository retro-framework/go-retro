package aggregates

import (
	"context"
	"errors"

	"github.com/leehambley/ls-cms/storage"
)

type AllowCreationOfNewIdentities struct {
	state Aggregate
}

// State returns a WidgetsApp from the Aggregate that everyone else
// wants to deal with, every Aggregate type must implement this.
func (cmd *AllowCreationOfNewIdentities) State() (*WidgetsApp, error) {
	if wa, ok := cmd.state.(*WidgetsApp); ok {
		return wa, nil
	} else {
		return &WidgetsApp{}, errors.New("can't cast")
	}
}

// AllowCreationOfNewIdentities is used to toggle the creation of new
// identites on (effectively enabling signup) it may be redundant in the
// case of systems that use a SSO such as active directory or OAuth. An
// application instance that has never had this called may default to
// "false" subject to how it was initialized.
func (cmd *AllowCreationOfNewIdentities) Apply(ctxt context.Context, sesh Session, repo storage.Depot) ([]string, error) {
	state, _ := cmd.State()

	// TODO: fix this to be sane, again
	// numIds, countable := repo.GetByDirname("identities").Len()
	// if !countable
	// 	return nil, errors.New("can't change application settings anonymously once identities exist")
	// }

	if state.allowCreateIdentities == true {
		return []string{}, nil
	}

	return []string{"AllowCreateIdentities"}, nil
}

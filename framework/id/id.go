package id

import "github.com/rs/xid"

func Xid() string {
	return xid.New().String()
}

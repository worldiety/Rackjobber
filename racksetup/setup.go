// Package racksetup includes functions to setup the master Repository
package racksetup

import (
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/gitutil"
)

// Setup will setup the rackjobber configuration.
func Setup() error {
	return gitutil.SetupMaster()
}

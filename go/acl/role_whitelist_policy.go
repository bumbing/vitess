package acl

import (
	"errors"
	"flag"
	"net/http"

	"vitess.io/vitess/go/flagutil"
	"vitess.io/vitess/go/vt/log"
)

// Example usage:
// vtctld <other args> -security_policy role_whitelist -whitelisted_roles monitoring,debugging

var allowedRolesFlag flagutil.StringListValue

func init() {
	flag.Var(&allowedRolesFlag, "whitelisted_roles", "comma separated list of roles for the 'role_whitelist' security policy")
	RegisterPolicy("role_whitelist", RoleWhitelistPolicy{})
}

var errRoleWhitelist = errors.New("not allowed: administrative commands")

// RoleWhitelistPolicy grants the whitelisted roles to all users.
type RoleWhitelistPolicy struct{}

func (rwp RoleWhitelistPolicy) allowsRole(role string) bool {
	for _, allowedRole := range allowedRolesFlag {
		if role == allowedRole {
			return true
		}
	}
	return false
}

// CheckAccessActor disallows all actor access.
func (rwp RoleWhitelistPolicy) CheckAccessActor(actor, role string) error {
	if rwp.allowsRole(role) {
		return nil
	}
	log.Infof("%v denied role %v", actor, role)
	return errRoleWhitelist
}

// CheckAccessHTTP disallows all HTTP access.
func (rwp RoleWhitelistPolicy) CheckAccessHTTP(req *http.Request, role string) error {
	if rwp.allowsRole(role) {
		return nil
	}
	log.Infof("Unknown HTTP user denied role %v", role)
	return errRoleWhitelist
}

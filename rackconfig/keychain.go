// +build !darwin

package rackconfig

import "errors"

//getGitPassFromKeychain return nothin, as this implementations use is just to enable conditional compilation
func getGitPassFromKeychain(account, domain string) (string, string, error) {
	return "", "", errors.New("keychain only available on darwin")
}

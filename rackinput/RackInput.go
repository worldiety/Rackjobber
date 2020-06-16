// Package rackinput includes functions for managing user input
package rackinput

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

// InputAuth asks the user for a basic auth input
func InputAuth(domain string) http.BasicAuth {
	auth := http.BasicAuth{}

	auth.Username = AwaitTextInput("Authentication required for " + domain + "\nUsername:")
	auth.Password = AwaitPasswordInput("Password:")

	return auth
}

//AwaitTextInput returns the user input
func AwaitTextInput(label string) string {
	fmt.Println(label)

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	return scanner.Text()
}

//AwaitPasswordInput returns the user input, which is hidden in the console
func AwaitPasswordInput(label string) string {
	fmt.Println(label)

	password, _ := terminal.ReadPassword(int(syscall.Stdin)) //nolint, conversion necessary for Windows compilation
	encPassword := hex.EncodeToString(password)

	return encPassword
}

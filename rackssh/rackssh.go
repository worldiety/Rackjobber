// Package rackssh includes functions to establish and use a ssh connection to a server running shopware
package rackssh

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	sshrw "github.com/mosolovsa/go_cat_sshfilerw"
	"golang.org/x/crypto/ssh"

	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/gitutil"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackshop"
)

// ConnectToRemote will establish a connection to a specific host with the given config.
func ConnectToRemote(host string, config *ssh.ClientConfig) (*ssh.Client, error) {
	address := host + ":22"
	return ssh.Dial("tcp", address, config)
}

// CheckConnection will check if a connection to a host is possible with the given config.
func CheckConnection(host string, config *ssh.ClientConfig) bool {
	client, err := ConnectToRemote(host, config)
	if err != nil {
		fmt.Println(err)
		return false
	}

	defer func() {
		if err = client.Close(); err != nil {
			log.Printf("client.Close - error: %v\n", err)
		}
	}()

	return true
}

//GetRemoteFileFromShop loads a remote file from a shop and returns its local path
func GetRemoteFileFromShop(shop *rackshop.RackShop, file string) ([]byte, error) {
	fileDir := filepath.Join(shop.ShopwareDir, file)
	return readRemoteFileForShop(shop, fileDir)
}

//GetRemoteDirsFromShop returns a list of folders inside of the given dir and shop
func GetRemoteDirsFromShop(shop *rackshop.RackShop, dir string) []string {
	dirsDir := filepath.Join(shop.ShopwareDir, dir)
	return readRemoteDirsForShop(shop, dirsDir)
}

//readRemoteFileForShop returns the requested file of a given shop
func readRemoteFileForShop(shop *rackshop.RackShop, path string) ([]byte, error) {
	var buff bytes.Buffer

	c, err := connectToShop(shop)
	if err != nil {
		fmt.Println("Failed to connect to shop")
		return nil, err
	}
	defer c.Close()

	w := bufio.NewWriter(&buff)

	err = c.ReadFile(w, path)
	if err != nil {
		fmt.Println("Error on file read: ", err.Error())
		return nil, err
	}

	err = w.Flush()
	if err != nil {
		fmt.Printf("bufioWriter.Flush - error: %v\n", err)
	}

	return buff.Bytes(), nil
}

//readRemoteDirsForShop returns a list of folders inside of the given dir and shop
func readRemoteDirsForShop(shop *rackshop.RackShop, path string) []string {
	command := "cd " + path + " && ls -d */"

	conn, _ := ssh.Dial("tcp", shop.Address+":22", shop.GetRemoteConfig())

	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("conn.Close - error: %v\n", err)
		}
	}()

	session, _ := conn.NewSession()

	defer func() {
		if err := session.Close(); err != nil {
			log.Printf("session.Close - error: %v\n", err)
		}
	}()

	var b bytes.Buffer
	session.Stdout = &b

	err := session.Run(command)
	if err != nil {
		log.Fatalf("session.Run - error: %v\n", err)
	}

	return strings.Split(strings.ReplaceAll(b.String(), "/", ""), "\n")
}

//connectToShop returns a SSHClient for the given shop
func connectToShop(shop *rackshop.RackShop) (*sshrw.SSHClient, error) {
	return sshrw.NewSSHclt(shop.Address+":22", shop.GetRemoteConfig())
}

//RunRemoteCommandInShop runs a command on the remote machine
func RunRemoteCommandInShop(command string, shop *rackshop.RackShop) {
	conn, err := ssh.Dial("tcp", shop.Address+":22", shop.GetRemoteConfig())
	if err != nil {
		log.Fatalf("Could not connect to remote machine: %v\n", err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("conn.Close - error: %v\n", err)
		}
	}()

	session, err := conn.NewSession()
	if err != nil {
		log.Fatalf("Error creating new SSH session: %v\n", err)
	}

	defer func() {
		if err := session.Close(); err != nil && !strings.Contains(err.Error(), "EOF") {
			log.Printf("session.Close - error: %v\n", err)
		}
	}()

	err = session.Run(command)
	if err != nil {
		log.Fatalf("session.Run - error: %v\n", err)
	}
}

//CloneGitToRemoteShop clones a GIT repo to the remote server of a shop
func CloneGitToRemoteShop(shop *rackshop.RackShop, pluginName string, url string, version string) {
	fmt.Println("Cloning " + pluginName + " to " + shop.Name + ".")

	filledURL := gitutil.GetURLWithAuth(url)
	remoteClonePath := filepath.Join(shop.ShopwareDir, "custom", "plugins", pluginName)

	command := ""
	if version != "" {
		command = "git clone --single-branch --branch " + version + " " + filledURL + " " + remoteClonePath
	} else {
		command = "git clone " + filledURL + " " + remoteClonePath
	}

	RunRemoteCommandInShop(command, shop)
}

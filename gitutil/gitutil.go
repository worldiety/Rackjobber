// Package gitutil includes functions to connect to and manage GIT-based Repositories
package gitutil

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/src-d/go-git.v4"

	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/fileutil"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackconfig"
)

// MasterRepo returns the default RackRepo for RackJobber
func MasterRepo() string {
	return "https://github.com/worldiety/Rackspecs.git"
}

// Fetch will fetch a repository and will ask for authentication if necessary
func Fetch(repo git.Repository, domain string, opts git.FetchOptions) error {
	err := repo.Fetch(&opts)

	if checkForAuthenticationError(err) {
		auth := rackconfig.GetGITAuth(domain)

		authOpts := git.FetchOptions{
			RemoteName: opts.RemoteName,
			RefSpecs:   opts.RefSpecs,
			Depth:      opts.Depth,
			Auth:       &auth,
			Progress:   opts.Progress,
			Tags:       opts.Tags,
			Force:      opts.Force,
		}

		err = repo.Fetch(&authOpts)
		if checkForAuthenticationError(err) {
			log.Fatalln("Username or Password are wrong - Aborting")

			err = errors.New("Username or password are wrong")
		}
	}

	return err
}

// Clone will clone a repository to a specified path
func Clone(path string, opts git.CloneOptions) (*git.Repository, error) {
	repository, err := git.PlainClone(path, false, &opts)

	if checkForAuthenticationError(err) {
		err = os.RemoveAll(path + ".git")
		if err != nil {
			log.Printf("Failed to remove .git directory: %v\n", err)
		}

		domain := strings.Split(opts.URL, "/")[2]
		auth := rackconfig.GetGITAuth(domain)

		authOpts := git.CloneOptions{
			URL:               opts.URL,
			Auth:              &auth,
			RemoteName:        opts.RemoteName,
			ReferenceName:     opts.ReferenceName,
			SingleBranch:      opts.SingleBranch,
			NoCheckout:        opts.NoCheckout,
			Depth:             opts.Depth,
			RecurseSubmodules: opts.RecurseSubmodules,
			Progress:          opts.Progress,
			Tags:              opts.Tags,
		}

		repository, err = git.PlainClone(path, false, &authOpts)
		if checkForAuthenticationError(err) {
			log.Fatalln("Username or Password are wrong - Aborting")

			err = errors.New("Username or password are wrong")
		}
	}

	if err != nil {
		err = os.Remove(path)
		if err != nil {
			log.Printf("Failed to remove directory: %v\n", err)
		}
	}

	return repository, err
}

// Update will update a git worktree and will ask for authentication if necessary
func Update(worktree git.Worktree, domain string, opts git.PullOptions) error {
	err := worktree.Pull(&opts)

	if checkForAuthenticationError(err) {
		auth := rackconfig.GetGITAuth(domain)

		authOpts := git.PullOptions{
			RemoteName:        opts.RemoteName,
			ReferenceName:     opts.ReferenceName,
			SingleBranch:      opts.SingleBranch,
			Depth:             opts.Depth,
			Auth:              &auth,
			RecurseSubmodules: opts.RecurseSubmodules,
			Progress:          opts.Progress,
			Force:             opts.Force,
		}

		err = worktree.Pull(&authOpts)

		if checkForAuthenticationError(err) {
			log.Fatalln("Username or Password are wrong - Aborting")

			err = errors.New("Username or password are wrong")
		}
	}

	return err
}

// Push will push the current worktree of a repository to its origin and will ask for authentication if necessary
func Push(repo git.Repository, domain string, opts git.PushOptions) error {
	err := repo.Push(&opts)
	if checkForAuthenticationError(err) {
		auth := rackconfig.GetGITAuth(domain)

		authOpts := git.PushOptions{
			RemoteName: opts.RemoteName,
			RefSpecs:   opts.RefSpecs,
			Auth:       &auth,
			Progress:   opts.Progress,
		}

		err = repo.Push(&authOpts)

		if checkForAuthenticationError(err) {
			log.Fatalln("Username or Password are wrong - Aborting")

			err = errors.New("Username or password are wrong")
		}
	}

	return err
}

func checkForAuthenticationError(err error) bool {
	if err != nil && strings.Contains(err.Error(), "authentication required") {
		return true
	}

	return false
}

// OpenCurrentDirGit will try to open a git repository from the current folder
func OpenCurrentDirGit() (*git.Repository, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return GetRepoFromLocalDir(dir)
}

// GetRepoFromLocalDir will try to open a git repository from a specified folder
func GetRepoFromLocalDir(directory string) (*git.Repository, error) {
	repo, err := git.PlainOpen(directory)
	if err != nil {
		return nil, err
	}

	return repo, err
}

// GetURLForRepo will extract the URL for this repository.
// Note that it will only consider the url of the first Remote entry
func GetURLForRepo(repo git.Repository) (*string, error) {
	remotes, err := repo.Remotes()
	if err != nil {
		return nil, err
	}

	if len(remotes) > 0 {
		first := remotes[0]
		config := first.Config()
		url := config.URLs[0]

		return &url, nil
	}

	return nil, errors.New("no remotes configured for repository")
}

//ReinstallMaster deletes the Rackspec Master-Repository and clones it again
func ReinstallMaster() error {
	fmt.Println("updating Master Repository")

	path, err := getRepoStorePath()
	if err != nil {
		log.Printf("failed to retrieve RepoStorePath: %v\n", err)
		return err
	}

	err = fileutil.DeleteDirectory(*path)
	if err != nil {
		log.Printf("Failed to delete Master Repository. Aborting reinstall: %v\n", err)
		return err
	}

	err = SetupMaster()
	if err != nil {
		return err
	}

	fmt.Println("successfully updated Master Repository")

	return nil
}

//SetupMaster clones the Master Repository of the rackjobber configuration to the executing device
func SetupMaster() error {
	repoPath, err := getRepoStorePath()
	if err != nil {
		log.Fatalf("GetRepoPath - error: %v\n", err)
		return err
	}

	if repoPath == nil {
		log.Fatalln("RepoPath is unexpectedly nil")
		return errors.New("Unexpected nil value")
	}

	masterPath := filepath.Join(*repoPath, "master")

	_, err = Clone(masterPath, git.CloneOptions{
		URL:      MasterRepo(),
		Progress: os.Stdout,
	})

	if err != nil {
		log.Fatalf("PlainClone - error: %v\n", err)
	}

	return err
}

// GetRepoStorePath returns the path to the repository store directory.
func getRepoStorePath() (*string, error) {
	exPath, err := fileutil.GetAppFolderPath()
	if err != nil {
		return nil, err
	}

	repoStorePath := filepath.Join(*exPath, "repos")

	return &repoStorePath, nil
}

// GetURLWithAuth cuts the given URL to its domain and adds authentication to it
func GetURLWithAuth(url string) string {
	domain := CutURLToDomain(url)
	auth := rackconfig.GetGITAuth(domain)

	return strings.Replace(url, "https://", "https://"+auth.Username+":"+auth.Password+"@", -1)
}

// GetHashOfLastCommit retrieves the Hash-value of the latest commit of a given repository
func GetHashOfLastCommit(url, version string) (*string, error) {
	filledURL := GetURLWithAuth(url)
	command := *exec.Command("git", "ls-remote", "--tags", filledURL) //nolint, as it is a listing command

	var out bytes.Buffer

	command.Stdout = &out

	err := command.Run()
	if err != nil {
		log.Fatalf("Failed to run 'git ls-remote' command, err: %v\n", err)
	}

	var hash string

	if strings.Contains(out.String(), version+"^{}") {
		hash = cutRefs(out.String(), version)
	} else {
		hash = cutRefsClean(out.String(), version)
	}

	return &hash, nil
}

//CutURLToDomain cuts a given URL and returns the base domain
func CutURLToDomain(url string) string {
	cut := strings.Split(url, "/")[2]
	return cut
}

func cutRefsClean(refs, version string) string {
	cut := strings.Split(refs, "	refs/tags/"+version)[0]
	cuts := strings.Split(cut, "\n")
	cut = cuts[len(cuts)-1]
	cut = strings.TrimSpace(cut)

	return cut
}

func cutRefs(refs, version string) string {
	cut := strings.Split(refs, version+"^{}")[0]
	cut = strings.Split(cut, version)[1]
	cut = strings.Split(cut, "	")[0]
	cut = strings.TrimSpace(cut)

	return cut
}

// Package repository includes functions to manage GIT-based repositories
package repository

import (
	"encoding/xml"
	"errors"
	"strings"
	"time"

	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/fileutil"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/gitutil"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackconfig"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackplugin"
	"gitlab.worldiety.net/worldiety/customer/wdy/libriety/shopware/rackjobber/rackspec"

	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// GetRepoStorePath returns the path to the repository store directory.
func GetRepoStorePath() (*string, error) {
	exPath, err := fileutil.GetAppFolderPath()
	if err != nil {
		return nil, err
	}

	repoStorePath := filepath.Join(*exPath, "repos")

	return &repoStorePath, nil
}

// GetSpecificRepoPath will return the path to a specific Repository Directory
func GetSpecificRepoPath(repoName string) (*string, error) {
	repoStorePath, err := GetRepoStorePath()
	if err != nil {
		return nil, err
	}

	specificPath := filepath.Join(*repoStorePath, repoName)

	return &specificPath, nil
}

// CreateRepoStoreDir creates the repository directory for RackJobber
func CreateRepoStoreDir() error {
	repoStorePath, err := GetRepoStorePath()
	if err != nil {
		return err
	}

	exists, err := fileutil.ObjectExists(*repoStorePath)
	if err != nil {
		return err
	}

	if exists {
		log.Fatalln("Repo Directory already exists - Aborting")
		return errors.New("Directory already exists")
	}

	err = os.Mkdir(*repoStorePath, os.ModePerm)

	return err
}

// UpdateRepos will update all repositories
func UpdateRepos() error {
	repoStorePath, err := GetRepoStorePath()
	if err != nil {
		return err
	}

	directories, err := ioutil.ReadDir(*repoStorePath)
	if err != nil {
		return err
	}

	for _, dir := range directories {
		if dir.IsDir() {
			err := updateRepo(dir.Name())
			if err != nil {
				log.Printf("Failed to update repo %v: %v\n", dir.Name(), err)
				return err
			}
		}
	}

	fmt.Println("Successfully updated repositories")

	return nil
}

// ListRepos will make a map of all Repos connected to RackJobber with their Git URL
func ListRepos() (*map[string]string, error) {
	repos := make(map[string]string)

	repoStorePath, err := GetRepoStorePath()
	if err != nil {
		return nil, err
	}

	directories, err := ioutil.ReadDir(*repoStorePath)

	for _, dir := range directories {
		if dir.IsDir() {
			dirPath := filepath.Join(*repoStorePath, dir.Name())

			repo, err := git.PlainOpen(dirPath)
			if err != nil {
				return nil, err
			}

			remotes, err := repo.Remotes()
			if err != nil {
				return nil, err
			}

			if len(remotes) > 0 {
				first := remotes[0]
				config := first.Config()
				url := config.URLs[0]

				repos[dir.Name()] = url
			} else {
				return nil, err
			}
		}
	}

	return &repos, err
}

func fetchRepo(repoName string) error {
	fmt.Printf("Start fetching repo with name: %v\n", repoName)

	repoPath, err := GetSpecificRepoPath(repoName)
	if err != nil {
		log.Fatalf("GetSpecificRepoPath - error: %v\n", err)
		return err
	}

	repo, err := git.PlainOpen(*repoPath)
	if err != nil {
		log.Fatalf("PlainOpen - error: %v\n", err)
		return err
	}

	domain := getRepoDomain(repoName)

	err = gitutil.Fetch(*repo, domain, git.FetchOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	})
	if err != nil {
		return err
	}

	return nil
}

func updateRepo(repoName string) error {
	err := fetchRepo(repoName)
	if err != nil {
		log.Printf("FetchRepo - error: %v\n", err)
		return err
	}

	fmt.Printf("Start updating repo with name: %v\n", repoName)

	repoPath, err := GetSpecificRepoPath(repoName)
	if err != nil {
		log.Printf("GetSpecificRepoPath - error: %v\n", err)
		return err
	}

	repo, err := git.PlainOpen(*repoPath)
	if err != nil {
		log.Printf("PlainOpen - error: %v\n", err)
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		log.Printf("Worktree - error: %v\n", err)
		return err
	}

	domain := getRepoDomain(repoName)

	err = gitutil.Update(*worktree, domain, git.PullOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	})
	if err != nil {
		log.Printf("Update - error: %v\n", err)
		return err
	}

	ref, err := repo.Head()
	if err != nil {
		log.Printf("Head - error: %v\n", err)
		return err
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		log.Printf("CommitObject - error: %v\n", err)
		return err
	}

	fmt.Printf("Updated repo with name: %v\n", repoName)
	fmt.Printf("Commit message: %v\n", commit)

	return nil
}

// PushSpecToRepo pushes a rackspec to a specific repository
func PushSpecToRepo(repoName string) error {
	err := updateRepo(repoName)
	if err != nil {
		return err
	}

	srcRackSpecPath, err := rackspec.FindRackSpecInCurrentDir()
	if err != nil {
		return err
	}

	rackSpec, err := rackspec.UnmarshalRackSpec(*srcRackSpecPath)
	if err != nil {
		return err
	}

	if err = versionsCorrespond(rackSpec); err != nil {
		return err
	}

	pluginName := rackSpec.Name

	repoPath, err := GetSpecificRepoPath(repoName)
	if err != nil {
		return err
	}

	if err = fileExists(*repoPath); err != nil {
		return err
	}

	pluginDir := filepath.Join(*repoPath, pluginName)

	err = fileutil.CreateDirIfNotExistant(pluginDir)
	if err != nil {
		return err
	}

	versionDir := filepath.Join(pluginDir, rackSpec.Version)

	err = fileutil.CreateDirIfNotExistant(versionDir)
	if err != nil {
		return err
	}

	dstRackSpecPath := filepath.Join(versionDir, filepath.Base(*srcRackSpecPath))

	err = fileutil.CopyFile(*srcRackSpecPath, dstRackSpecPath)
	if err != nil {
		return err
	}

	entryPath, err := filepath.Rel(*repoPath, dstRackSpecPath)
	if err != nil {
		return err
	}

	err = commitYml(entryPath, repoName, rackSpec.Name, rackSpec.Version, versionDir)

	return err
}

func commitYml(entryPath, repoName, name, version, versionDir string) error {
	commitMsg := fmt.Sprintf("Add %v_rackspec.yaml for Version %v to repo", name, version)

	err := addFileToGit(entryPath, repoName, commitMsg)
	if err != nil {
		remerr := os.RemoveAll(versionDir)
		if remerr != nil {
			log.Printf("os.Remove - error: %v\n", remerr)
		}

		return err
	}

	return nil
}

func fileExists(repoPath string) error {
	exists, err := fileutil.ObjectExists(repoPath)
	if err != nil {
		return nil
	}

	if !exists {
		log.Fatalf("Could not find repository at path: %v\n", repoPath)
		return errors.New("Repository does not exist")
	}

	return nil
}

func addFileToGit(filePath string, repoName string, commitMsg string) error {
	repoPath, err := GetSpecificRepoPath(repoName)
	if err != nil {
		return err
	}

	repo, err := git.PlainOpen(*repoPath)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	fmt.Println("File to add:")
	fmt.Println(filePath)

	_, err = worktree.Add(filePath)
	if err != nil {
		return err
	}

	status, err := worktree.Status()
	if err != nil {
		return err
	}

	fmt.Print("Current Worktree status: \n\n")
	fmt.Println(status)
	fmt.Println("Create Commit")

	err = commit(repo, worktree, commitMsg)
	if err != nil {
		return err
	}

	fmt.Println("Pushing commits:")

	domain := getRepoDomain(repoName)

	err = gitutil.Push(*repo, domain, git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	})
	if err != nil {
		return err
	}

	return nil
}

func commit(repo *git.Repository, worktree *git.Worktree, commitMsg string) error {
	commit, err := worktree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Rack Jobber",
			Email: "rack@jobber.org",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}

	obj, err := repo.CommitObject(commit)
	if err != nil {
		return err
	}

	fmt.Print("Created Commit: \n\n")
	fmt.Println(obj)

	return nil
}

// AddRepo will add a new rackspec repository to rackjobber
func AddRepo(name string, source string) error {
	repoStorePath, err := GetRepoStorePath()
	if err != nil {
		log.Fatalf("RepoStorePath - error: %v\n", err)
		return err
	}

	if repoStorePath == nil {
		log.Fatalln("RepoStorePath is unexpectedly nil")
		return errors.New("Unexpected nil value")
	}

	newRepoStorePath := filepath.Join(*repoStorePath, name)

	_, err = gitutil.Clone(newRepoStorePath, git.CloneOptions{
		URL:      source,
		Progress: os.Stdout,
	})

	if err != nil {
		log.Fatalf("PlainClone - error: %v\n", err)
	}

	return err
}

// RemoveRepo will remove a rackspec repository from rackjobber
func RemoveRepo(name string) error {
	repoStorePath, err := GetRepoStorePath()
	if err != nil {
		return err
	}

	repoToRemove := filepath.Join(*repoStorePath, name)

	err = os.RemoveAll(repoToRemove)

	return err
}

//getRepoDomain extracts and returns the domain of a repo
func getRepoDomain(name string) string {
	repoPath, _ := GetSpecificRepoPath(name)
	repo, _ := gitutil.GetRepoFromLocalDir(*repoPath)
	repoURL, _ := gitutil.GetURLForRepo(*repo)

	return strings.Split(*repoURL, "/")[2]
}

//originVersionTagExists checks if the version specified in the given rackspec exists in the corresponding git repo
func originVersionTagExists(rackSpec *rackspec.RackSpec) error {
	repo, err := git.Init(memory.NewStorage(), nil)
	if err != nil {
		return err
	}

	domain := gitutil.CutURLToDomain(rackSpec.Source.GIT)

	auth := rackconfig.GetGITAuth(domain)

	filledURL := strings.Replace(rackSpec.Source.GIT, "https://", "https://"+auth.Username+":"+auth.Password+"@", -1)

	remote, err := repo.CreateRemote(&config.RemoteConfig{
		Name: "rackjobber",
		URLs: []string{filledURL},
	})
	if err != nil {
		return err
	}

	references, err := remote.List(&git.ListOptions{})
	if err != nil {
		return err
	}

	hasTag := false

	for _, ref := range references {
		if strings.Contains(ref.String(), rackSpec.Version) {
			hasTag = true
		}
	}

	if !hasTag {
		err := "the version specified in the rackspec.yml does not correspond to any version-tag on the repository"
		return errors.New(err)
	}

	return nil
}

//versionsCorrespond checks if the version specified in the rackspec.yaml
// corresponds to the one specified in the plugin.xml
func versionsCorrespond(rackSpec *rackspec.RackSpec) error {
	if err := originVersionTagExists(rackSpec); err != nil {
		return err
	}

	currentDirectory, err := os.Getwd()
	if err != nil {
		return err
	}

	xmlpath := filepath.Join(currentDirectory, "plugin.xml")

	pluginxml, err := ioutil.ReadFile(xmlpath) //nolint, as the file is only being unmarshaled
	if err != nil {
		errMsg := "no plugin.xml found. If you were to push the rackspec now, "
		errMsg += "shopware would not recognize an available update for your plugin"

		return errors.New(errMsg)
	}

	plugin := &rackplugin.Plugin{}

	err = xml.Unmarshal(pluginxml, &plugin)
	if err != nil {
		return err
	}

	if rackSpec.Version == plugin.Version {
		return nil
	}

	errMsg := "version in the rackspec.yml does not correspond to the one in the plugin.xml."
	errMsg += "If you were to push the rackspec now, shopware would not recognize an available update for your plugin"

	return errors.New(errMsg)
}

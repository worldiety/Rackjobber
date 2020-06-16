// Package fileutil includes functions to read, write and work with files and directories
package fileutil

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ObjectExists checks if a directory or file exists at the specified path
func ObjectExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return true, err
}

// ReadFile will read out the data of a file and return it.
func ReadFile(path string) (*[]byte, error) {
	if err := ValidateFilepath(path); err != nil {
		log.Printf("ValidateFilepath - error: %v\n", err)
		return nil, err
	}

	file, err := os.Open(path) //nolint, as the file is validated above
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("file.Close - error: %v\n", err)
		}
	}()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

// CreateOrWriteFile creates a file if it does not exist, or overrides it.
func CreateOrWriteFile(path string, data []byte) error {
	exists, err := ObjectExists(path)
	if err != nil {
		return err
	}

	if exists {
		return OverrideFile(path, data)
	}

	return CreateFile(path, data)
}

// CreateFile will create a file at a specified path
func CreateFile(path string, data []byte) error {
	return ioutil.WriteFile(path, data, os.ModePerm)
}

// OverrideFile will override a file if it exists
func OverrideFile(path string, data []byte) error {
	err := os.Remove(path)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, os.ModePerm)
}

// GetAppPath returns the path of the current application
func getAppPath() (*string, error) {
	ex, err := os.Executable()
	if err != nil {
		return nil, err
	}

	exPath := filepath.Dir(ex)

	return &exPath, nil
}

// GetAppFolderPath will return the folder path of the executable where config data is stored.
func GetAppFolderPath() (*string, error) {
	exPath, err := getAppPath()
	if err != nil {
		return nil, err
	}

	folderPath := filepath.Join(*exPath, "rackresource")

	return &folderPath, nil
}

// CopyFile will copy a file from a specified source to a specified destination
func CopyFile(src string, dst string) error {
	// Check if the source exists
	sourceFile, err := os.Stat(src)
	if err != nil {
		return err
	}

	// check if the source is a file
	if !sourceFile.Mode().IsRegular() {
		return errors.New("src is no file")
	}

	source, err := os.Open(src) //nolint 
	if err != nil {
		return err
	}

	defer func() {
		if err := source.Close(); err != nil {
			log.Printf("source.Close - error: %v\n", err)
		}
	}()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer func() {
		if err := destination.Close(); err != nil {
			log.Printf("destination.Close - error: %v\n", err)
		}
	}()

	_, err = io.Copy(destination, source)

	return err
}

// CreateDirIfNotExistant will create a directory when it does not exist.
func CreateDirIfNotExistant(dir string) error {
	exists, err := ObjectExists(dir)
	if err != nil {
		return err
	}

	if !exists {
		err := os.Mkdir(dir, os.ModePerm)
		return err
	}

	return nil
}

//DeleteDirectory deletes a given directory, as well as all children
func DeleteDirectory(dir string) error {
	fmt.Printf("deleting directory: \n")

	err := os.RemoveAll(dir)
	if err != nil {
		log.Printf("Failed to delete directory: %v\n", err)
		return err
	}

	fmt.Printf("successfully deleted directory: \n")

	return nil
}

// ValidateFilepath ckecks if the given filepath lies within rackjobbers directory structure
func ValidateFilepath(filePath string) error {
	ex, err := os.Executable()
	if err != nil {
		return err
	}

	exPath := filepath.Dir(ex)
	if !strings.Contains(filePath, exPath) {
		return errors.New("provided filepath is outside of rackjobbers boundaries")
	}

	return nil
}

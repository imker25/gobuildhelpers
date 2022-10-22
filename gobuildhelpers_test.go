// Copyright 2022 by tobi@backfrak.de. All
// rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the
// LICENSE file.

package gobuildhelpers

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

const baseDir = "tmp"

func TestListContains(t *testing.T) {
	list := []string{"a", "b", "c"}

	if listContains(list, "1") == true {
		t.Errorf("The list '%s' contains the string '1'", list)
	}

	if listContains(list, "b") == false {
		t.Errorf("The list '%s' not contains the string 'b'", list)
	}

	list = append(list, "1")

	if listContains(list, "1") == false {
		t.Errorf("The list '%s' not contains the string '1'", list)
	}

	if listContains(list, "z") == true {
		t.Errorf("The list '%s' contains the string 'z'", list)
	}
}

func TestGetGitHash(t *testing.T) {
	gitHash, err := GetGitHash(".")

	if err != nil {
		t.Errorf("Got error '%s' but expected none", err.Error())
	}

	if gitHash == "" {
		t.Errorf("Got an empty string, but expected some content")
	}

	rootDir := "/"
	if runtime.GOOS == "windows" {
		rootDir = "\\"
	}
	gitHash, err = GetGitHash(rootDir)

	if err == nil {
		t.Errorf("Expected an error, but got none")
	}

	if gitHash != "" {
		t.Errorf("Expected '', but got '%s'", gitHash)
	}
}

func TestGetGitHeight(t *testing.T) {
	gitHeight, err := GetGitHeight("VersionMaster.txt", ".")

	if err != nil {
		t.Errorf("Got error '%s' but expected none", err.Error())
	}

	if gitHeight < 0 {
		t.Errorf("Got a height of '-1', but expected '0' or grater")
	}

	gitHeight, err = GetGitHeight("not_existing_file", ".")

	if err == nil {
		t.Errorf("Got no error, but expected one")
	}

	if gitHeight != -1 {
		t.Errorf("Expected git height to be '-1', but is '%d'", gitHeight)
	}
}

func TestRemovePaths(t *testing.T) {
	if err := createTmpDirs(); err != nil {
		t.Errorf("Got error while test preperation")
	}

	err := RemovePaths([]string{baseDir})

	if err != nil {
		t.Errorf("Got error '%s' but expected none", err.Error())
	}

	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		fmt.Println("OK")
	} else {
		t.Errorf("The path '%s' was not removed as expected", baseDir)
	}
}

func TestEnsureDirectoryExists(t *testing.T) {

	err := EnsureDirectoryExists(baseDir)
	if err != nil {
		t.Errorf("Got error '%s' but expected none", err.Error())
	}

	if _, err := os.Stat(baseDir); os.IsNotExist(err) {
		t.Errorf("The path '%s' was not created as expected", baseDir)
	}

	RemovePaths([]string{baseDir})

}

func TestZipFolders(t *testing.T) {
	createTmpDirs()
	mydir1 := filepath.Join(baseDir, "myDir1")
	mydir2 := filepath.Join(baseDir, "myDir2")
	outFile := "out.zip"

	err := ZipFolders([]string{mydir1, mydir2}, outFile)
	if err != nil {
		t.Errorf("Got error '%s' but expected none", err.Error())
	}
	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		t.Errorf("The path '%s' was not created as expected", outFile)
	}

	if errRem := RemovePaths([]string{outFile, baseDir}); errRem != nil {
		t.Errorf("Got error '%s' but expected none", errRem.Error())
	}

}

func createTmpDirs() error {

	if err := EnsureDirectoryExists(baseDir); err != nil {
		return err
	}
	mydir1 := filepath.Join(baseDir, "myDir1")
	if err := EnsureDirectoryExists(mydir1); err != nil {
		return err
	}
	mydir2 := filepath.Join(baseDir, "myDir2")
	if err := EnsureDirectoryExists(mydir2); err != nil {
		return err
	}
	mySubDir1 := filepath.Join(mydir1, "subDir1")
	if err := EnsureDirectoryExists(mySubDir1); err != nil {
		return err
	}
	mySubDir2 := filepath.Join(mydir1, "subDir2")
	if err := EnsureDirectoryExists(mySubDir2); err != nil {
		return err
	}

	file1 := filepath.Join(mydir1, "file1.txt")
	if err := os.WriteFile(file1, []byte("some content 1"), 0644); err != nil {
		return err
	}

	file2 := filepath.Join(mySubDir2, "file2.txt")
	if err := os.WriteFile(file2, []byte("some content 2"), 0644); err != nil {
		return err
	}
	return nil
}

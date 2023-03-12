// Copyright 2022 by tobi@backfrak.de. All
// rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the
// LICENSE file.

package gobuildhelpers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

	if !PathExists(outFile) {
		t.Errorf("The zip file '%s' does not exit", outFile)
	}

	if errRem := RemovePaths([]string{outFile}); errRem != nil {
		t.Errorf("Got error '%s' but expected none", errRem.Error())
	}

	err = ZipFolders([]string{mydir1}, filepath.Join(".", "testdata", "testResultConverter", "not-existing-dir", outFile))
	if err == nil {
		t.Errorf("Got no error, but expected one")
	}

	if errRem := RemovePaths([]string{outFile, baseDir}); errRem != nil {
		t.Errorf("Got error '%s' but expected none", errRem.Error())
	}

}

func TestInstallTestConvert(t *testing.T) {
	err := InstallTestConverter(filepath.Join(".", "testdata", "testResultConverter"))
	if err != nil {
		t.Errorf("Got error '%s' but expected none", err.Error())
	}

	err = InstallTestConverter(filepath.Join(".", "testdata", "testResultConverter", "not-existing-dir"))
	if err == nil {
		t.Errorf("Got no error but expected one")
	}
}

func TestFindFolderToBuild(t *testing.T) {
	dirs, err := FindPackagesToBuild(filepath.Join(".", "testdata", "testProject"))
	if err != nil {
		t.Errorf("Got error '%s', but expected none", err.Error())
	}

	if len(dirs) != 1 {
		t.Errorf("Expected '1' folder to build, but got '%d'", len(dirs))
	}

	dirs, err = FindPackagesToBuild(filepath.Join(".", "testdata", "no.go"))
	if err != nil {
		t.Errorf("Got the error '%s', but expected none", err.Error())
	}
	if len(dirs) != 0 {
		t.Errorf("Expected '0' folders to build, but got '%d'", len(dirs))
	}

}

func TestFindFoldersToTest(t *testing.T) {
	dirs, err := FindPackagesToTest(filepath.Join(".", "testdata", "testProject"))
	if err != nil {
		t.Errorf("Got error '%s', but expected none", err.Error())
	}

	if len(dirs) != 1 {
		t.Errorf("Expected '1' folder to test, but got '%d'", len(dirs))
	}

	dirs, err = FindPackagesToTest(filepath.Join(".", "testdata", "no.go"))
	if err != nil {
		t.Errorf("Got the error '%s', but expected none", err.Error())
	}
	if len(dirs) != 0 {
		t.Errorf("Expected '0' folders to test, but got '%d'", len(dirs))
	}

}

func TestBuild(t *testing.T) {
	RemovePaths([]string{baseDir})

	dirs, err := FindPackagesToBuild(filepath.Join(".", "testdata", "testProject"))
	if err != nil {
		t.Errorf("Got error '%s', but expected none", err.Error())
	}

	if len(dirs) != 1 {
		t.Errorf("Expected '1' folder to build, but got '%d'", len(dirs))
	}
	workDir, errWorkDir := os.Getwd()
	if errWorkDir != nil {
		t.Errorf("Got error '%s', but expected none", errWorkDir.Error())
	}
	outDir := filepath.Join(workDir, baseDir)
	errBuild := BuildFolders(dirs, outDir, "")
	if errBuild != nil {
		t.Errorf("Got error '%s', but expected none", errBuild.Error())
	}

	binOutPath := filepath.Join(workDir, baseDir, filepath.Base(dirs[0]))
	if runtime.GOOS == "windows" {
		binOutPath = fmt.Sprintf("%s.exe", binOutPath)
	}

	if !PathExists(binOutPath) {
		t.Errorf("Executable '%s', was not created", binOutPath)
	}

	cmd := exec.Command(binOutPath)
	outPutBytes, err := cmd.Output()
	if err != nil {
		t.Errorf("Got error '%s', but expected none", err.Error())
	}

	if strings.TrimSpace(string(outPutBytes)) != "The result is '3'" {
		t.Errorf("The executable does not return the expected string, it returns '%s'", strings.TrimSpace(string(outPutBytes)))
	}

	errBuild = BuildFolders(dirs, baseDir, "a=123")
	if errBuild != nil {
		t.Errorf("Got error '%s', but expected none", errBuild.Error())
	}

	errBuild = BuildFolders([]string{filepath.Join(".", "testdata", "no.go")}, baseDir, "")
	if errBuild == nil {
		t.Errorf("Got no error, but expected one")
	}
}

func TestPathExists(t *testing.T) {

	workDir, errWorkDir := os.Getwd()
	if errWorkDir != nil {
		t.Errorf("Got error '%s', but expected none", errWorkDir.Error())
	}

	notExist := filepath.Join(".", baseDir, "not-existing-dir", "not-existing-file")
	if PathExists(notExist) {
		t.Errorf("Found file '%s', but should not", notExist)
	}

	if !PathExists(workDir) {
		t.Errorf("Path '%s', does not exist", workDir)
	}
}

func TestTestExecution(t *testing.T) {

	RemovePaths([]string{baseDir})

	dirs, err := FindPackagesToTest(filepath.Join(".", "testdata", "testProject"))
	if err != nil {
		t.Errorf("Got error '%s', but expected none", err.Error())
	}

	if len(dirs) != 1 {
		t.Errorf("Expected '1' folder to build, but got '%d'", len(dirs))
	}

	errTests := RunTestFolders(dirs, baseDir, "TestResult.log")
	if len(errTests) != 0 {
		t.Errorf("Got error '%s', but expected none", errTests[0].Error())
	}

	testLog := filepath.Join(".", baseDir, "TestResult.log")
	if !PathExists(testLog) {
		t.Errorf("Test output file '%s' does not exist", testLog)
	}

	errTests = RunTestFolders([]string{filepath.Join(".", "testdata", "no.go")}, baseDir, "TestResult.log")
	if len(errTests) != 1 {
		t.Errorf("Got no error, but expected one")
	}
}

func TestTestExecutionEarlyExit(t *testing.T) {

	RemovePaths([]string{baseDir})

	dirs, err := FindPackagesToTest(filepath.Join(".", "testdata", "testProject"))
	if err != nil {
		t.Errorf("Got error '%s', but expected none", err.Error())
	}

	if len(dirs) != 1 {
		t.Errorf("Expected '1' folder to build, but got '%d'", len(dirs))
	}

	errTests := RunTestFoldersEarlyExit(dirs, baseDir, "TestResult.log")
	if errTests != nil {
		t.Errorf("Got error '%s', but expected none", errTests.Error())
	}

	testLog := filepath.Join(".", baseDir, "TestResult.log")
	if !PathExists(testLog) {
		t.Errorf("Test output file '%s' does not exist", testLog)
	}

	errTests = RunTestFoldersEarlyExit([]string{filepath.Join(".", "testdata", "no.go")}, baseDir, "TestResult.log")
	if errTests == nil {
		t.Errorf("Got no error, but expected one")
	}
}

func TestTestCoverage(t *testing.T) {
	RemovePaths([]string{baseDir})

	dirs, err := FindPackagesToTest(filepath.Join(".", "testdata", "testProject"))
	if err != nil {
		t.Errorf("Got error '%s', but expected none", err.Error())
	}

	if len(dirs) != 1 {
		t.Errorf("Expected '1' folder to build, but got '%d'", len(dirs))
	}

	errTests := CoverTestFolders(dirs, baseDir, "TestCover.log")
	if errTests != nil {
		t.Errorf("Got error '%s', but expected none", errTests.Error())
	}

	testLog := filepath.Join(".", baseDir, "TestCover.log")
	if !PathExists(testLog) {
		t.Errorf("Test output file '%s' does not exist", testLog)
	}

	errTests = CoverTestFolders([]string{filepath.Join(".", "testdata", "no.go")}, baseDir, "TestCover.log")
	if errTests == nil {
		t.Errorf("Got no error, but expected one")
	}
}

func TestTestCovertion(t *testing.T) {
	RemovePaths([]string{baseDir})

	testConvWorkDir := filepath.Join(".", "testdata", "testResultConverter")
	InstallTestConverter(testConvWorkDir)

	dirs, err := FindPackagesToTest(filepath.Join(".", "testdata", "testProject"))
	if err != nil {
		t.Errorf("Got error '%s', but expected none", err.Error())
	}

	if len(dirs) != 1 {
		t.Errorf("Expected '1' folder to build, but got '%d'", len(dirs))
	}

	errTests := RunTestFolders(dirs, filepath.Join(".", baseDir), "TestResult.log")
	if len(errTests) != 0 {
		t.Errorf("Got error '%s', but expected none", errTests[0].Error())
	}

	errConv := ConvertTestResults(filepath.Join("..", "..", baseDir, "TestResult.log"), filepath.Join("..", "..", baseDir, "TestResult.xml"), testConvWorkDir)
	if errConv != nil {
		t.Errorf("Got error '%s', but expected none", err.Error())
	}

	if !PathExists(filepath.Join(".", baseDir, "TestResult.xml")) {
		t.Errorf("The converted file '%s' does not exist", filepath.Join(".", baseDir, "TestResult.xml"))
	}

	errConv = ConvertTestResults(filepath.Join(".", baseDir, "TestResult.log"), filepath.Join("..", "..", baseDir, "TestResult.xml"), testConvWorkDir)
	if errConv == nil {
		t.Errorf("Got no error, but expected one")
	}
}

func TestReadOSDistribution(t *testing.T) {

	if runtime.GOOS == "linux" {
		dist, err := ReadOSDistribution()
		if err != nil {
			t.Errorf("Got error '%s', but expected none", err.Error())
		}
		if dist == "" {
			t.Errorf("Got an empty string instead of a distribution name")
		}
		fmt.Println(fmt.Sprintf("Distribution: '%s'", dist))
	}

	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		dist, err := ReadOSDistribution()
		if dist != "" {
			t.Errorf("Expected an empty string but got '%s'", dist)
		}
		if err == nil {
			t.Errorf("Got no error but expected one")

			return
		}

		switch err.(type) {
		case *OsNotSupportedByThisMethod:
			fmt.Fprintln(os.Stdout, "OK")
		default:
			t.Errorf("Got error '%s' type, but expected '*OsNotSupportedByThisMethod'", err.Error())
		}

		errMsg := err.Error()
		if !strings.Contains(errMsg, "ReadOSDistribution") {
			t.Errorf("Expected '%s' to contain 'ReadOSDistribution'", errMsg)
		}

		if !strings.Contains(errMsg, runtime.GOOS) {
			t.Errorf("Expected '%s' to contain '%s'", errMsg, runtime.GOOS)
		}
	}
}

func TestOsNotSupportedByThisMethod(t *testing.T) {
	os := "myOs"
	method := "myMethod"
	err := NewOsNotSupportedByThisMethod(os, method)

	if err.os != os {
		t.Errorf("NewOsNotSupportedByThisMethod.os is not the expected value '%s, but is '%s'", os, err.os)
	}

	if err.method != method {
		t.Errorf("NewOsNotSupportedByThisMethod.method is not the expected value '%s, but is '%s'", method, err.method)
	}

	if !strings.Contains(err.err, os) {
		t.Errorf("Error message '%s' does not contain '%s'", err.err, os)
	}

	if !strings.Contains(err.err, method) {
		t.Errorf("Error message '%s' does not contain '%s'", err.err, method)
	}

	if fmt.Sprintf("Error: %s", err.err) != err.Error() {
		t.Errorf("The error message '%s' is not the expected '%s'", err.Error(), fmt.Sprintf("Error: %s", err.err))
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

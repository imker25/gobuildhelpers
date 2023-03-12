// Copyright 2022 by tobi@backfrak.de. All
// rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the
// LICENSE file.

// gobuildhelpers - Go package with helper functions for go based CI/CD workflows
// May be used within https://magefile.org/ based build scripts
package gobuildhelpers

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type OsNotSupportedByThisMethod struct {
	err    string
	os     string
	method string
}

func (e *OsNotSupportedByThisMethod) Error() string { // Implement the Error Interface for the OsNotSupportedByThisMethod struct
	return fmt.Sprintf("Error: %s", e.err)
}

// NewOsNotSupportedByThisMethod- Get a new OsNotSupportedByThisMethod struct
func NewOsNotSupportedByThisMethod(os, method string) *OsNotSupportedByThisMethod {
	return &OsNotSupportedByThisMethod{fmt.Sprintf("The OS \"%s\" is not supported by the \"%s\" method", os, method), os, method}
}

// RemovePaths - Deletes the given paths recursively
// - paths: The list of directory or file path to delete
// It returns any error may occur or nil
func RemovePaths(paths []string) error {
	for _, path := range paths {
		err := os.RemoveAll(path)
		if err != nil {
			return err
		}
	}
	return nil
}

// ConvertTestResults - Converts a given 'go test' output log and converts the content into a junit xml result file
// It uses 'github.com/tebeka/go2xunit' to do so. You need to install this package before you can run this function.
// To install the converter you might want to use the 'InstallTestConverter' function
// - logPath: The path to the 'go test' output log to convert
// - xmlResult: The junit xml result output file
// - workDir: The directory this operation will run in. Usually the repository root directory
// It returns any error that may occur or nil
func ConvertTestResults(logPath, xmlResult, workDir string) error {
	xmlOutDir := filepath.Dir(xmlResult)
	if err := EnsureDirectoryExists(xmlOutDir); err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("Convert the test results %s to %s", logPath, xmlResult))
	cmd := exec.Command("go", "run", "github.com/tebeka/go2xunit", "-input", logPath, "-output", xmlResult)
	cmd.Dir = workDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	errConvert := cmd.Run()
	if errConvert != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Error during test result conversion: %s", errConvert.Error()))
		return errConvert
	}

	return nil
}

// InstallTestConverter - Install the  'github.com/tebeka/go2xunit' package used to convert test results in the 'ConvertTestResults' function
// - workDir: The directory the package will be installed. Might not the repository root
// It returns any error that may occur or nil
func InstallTestConverter(workDir string) error {
	cmd := exec.Command("go", "install", "-v", "github.com/tebeka/go2xunit@v1.4.10")
	cmd.Dir = workDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	errInst := cmd.Run()
	if errInst != nil {
		return errInst
	}

	return nil
}

// ZipFolders - Zips the given source folders recursively into the target zip file
// - sources: List of path to the folders to zip
// - target: The output zip file
// It returns any error that may occur or nil
func ZipFolders(sources []string, target string) error {
	fmt.Println(fmt.Sprintf("Zip %s into %s", sources, target))
	// 1. Create a ZIP file and zip.Writer
	f, err := os.Create(target)
	if err != nil {
		return err
	}
	defer f.Close()

	writer := zip.NewWriter(f)
	defer writer.Close()

	for _, source := range sources {

		if _, err := os.Stat(source); os.IsNotExist(err) {
			continue
		}
		// 2. Go through all the files of the source
		packSourceErr := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 3. Create a local file header
			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			// set compression
			header.Method = zip.Deflate

			// 4. Set relative path of a file as the header name
			header.Name, err = filepath.Rel(filepath.Dir(source), path)
			if err != nil {
				return err
			}
			if info.IsDir() {
				header.Name += "/"
			}

			// 5. Create writer for the file header and save content of the file
			headerWriter, err := writer.CreateHeader(header)
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(headerWriter, f)
			if err != nil {
				return err
			}

			return nil
		})

		if packSourceErr != nil {
			return packSourceErr
		}
	}

	return nil
}

// GetGitHash - Get the git hash currently checked out in the workDir
// - workDir: The directory this operation will run in. Usually the repository root directory
// It returns the git hash string and nil in case no error occur
// In case of error the error and an empty string is returned
func GetGitHash(workDir string) (string, error) {
	cmd := exec.Command("git", "describe", "--always", "--long", "--dirty")
	cmd.Dir = workDir
	cmd.Stderr = os.Stderr
	hash, err := cmd.Output()
	if err != nil {
		return "", err
	}
	hashStr := strings.TrimSpace(string(hash))
	return hashStr, nil
}

// GetGitHeight - Get the git height ( https://github.com/dotnet/Nerdbank.GitVersioning#what-is-git-height ) to the last change of versionFile
// - versionFile: The relative path (to workDir) of the file git height is calculated for
// - workDir: The directory this operation will run in. Usually the repository root directory
// It returns the git height number and nil in case no error occur
// In case of error the error and '-1' is returned
func GetGitHeight(versionFile, workDir string) (int, error) {
	cmd := exec.Command("git", "log", "--pretty=format:\"%H\"", "-n 1", "--follow", versionFile)
	cmd.Dir = workDir
	cmd.Stderr = os.Stderr
	lastChange, errLast := cmd.Output()
	if errLast != nil {
		return -1, errLast
	}
	lastChangeStr := strings.ReplaceAll(strings.TrimSpace(string(lastChange)), "\"", "")

	cmd = exec.Command("git", "log", "--pretty=format:\"%H\"", "-n 1")
	cmd.Dir = workDir
	cmd.Stderr = os.Stderr
	head, errHead := cmd.Output()
	if errHead != nil {
		return -1, errHead
	}

	headStr := strings.ReplaceAll(strings.TrimSpace(string(head)), "\"", "")

	cmd = exec.Command("git", "rev-list", "--count", lastChangeStr+".."+headStr)
	cmd.Dir = workDir
	cmd.Stderr = os.Stderr
	height, heightErr := cmd.Output()
	if heightErr != nil {
		return -1, heightErr
	}

	heightStr := strings.TrimSpace(string(height))
	heightInt, errCon := strconv.Atoi(heightStr)
	if errCon != nil {
		return -1, nil
	}

	return heightInt, nil
}

// CoverTestFolders - Runs 'go test -v -cover' on all given packages to test and creates a log file with the output
// Any package folder in the list should contain a go package with at least one '*_test.go' file
// - packagesToCover: List of directory path that contains '*_test.go' files test coverage should be measured
// - logDir: Path to the directory the log file is crated
// - logFileName: Name of the log file
// It returns any error that may occur or nil
func CoverTestFolders(packagesToCover []string, logDir, logFileName string) error {
	if err := EnsureDirectoryExists(logDir); err != nil {
		return err
	}

	logPath := filepath.Join(logDir, logFileName)
	logFile, errOpen := os.Create(logPath)
	if errOpen != nil {
		return errOpen
	}
	defer logFile.Close()

	for _, packToTest := range packagesToCover {

		fmt.Println(fmt.Sprintf("Measure test coverage for package '%s', logging to '%s'", packToTest, logPath))
		fmt.Println(fmt.Sprintf("Run in %s: %s %s %s %s >> %s", packToTest, "go", "test", "-v", "-cover", logPath))
		cmd := exec.Command("go", "test", "-v", "-cover")

		cmd.Dir = packToTest
		cmd.Stderr = logFile
		cmd.Stdout = logFile
		errTest := cmd.Run()
		if errTest != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("Error during coverage measurement of package '%s': %s", packToTest, errTest.Error()))
			return errTest
		}
	}

	return nil
}

// RunTestFolders - Runs 'go test -v -race' on linux and 'go test -v' on windows for all given packages to test
// Any package folder in the list should contain a go package with at least one '*_test.go' file
// All tests will be executed, even if a error occur in the package before, the next package's tests get executed
// - packagesToTest: List of directory path that contains '*_test.go' files to run
// - logDir: Path to the directory the log file is crated
// - logFileName: Name of the log file
// It returns any error that may occur or an empty list
func RunTestFolders(packagesToTest []string, logDir, logFileName string) []error {
	testErrors := []error{}

	if err := EnsureDirectoryExists(logDir); err != nil {
		return append(testErrors, err)
	}

	logPath := filepath.Join(logDir, logFileName)
	logFile, errOpen := os.Create(logPath)
	if errOpen != nil {
		return append(testErrors, errOpen)
	}
	defer logFile.Close()

	for _, packToTest := range packagesToTest {

		fmt.Println(fmt.Sprintf("Test package '%s', logging to '%s'", packToTest, logPath))
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			fmt.Println(fmt.Sprintf("Run in %s: %s %s %s >> %s", packToTest, "go", "test", "-v", logPath))
			cmd = exec.Command("go", "test", "-v")
		} else {
			fmt.Println(fmt.Sprintf("Run in %s: %s %s %s %s >> %s", packToTest, "go", "test", "-v", "-race", logPath))
			cmd = exec.Command("go", "test", "-v", "-race")
		}
		cmd.Dir = packToTest
		cmd.Stderr = logFile
		cmd.Stdout = logFile
		errTest := cmd.Run()
		if errTest != nil {
			fmt.Fprintln(os.Stderr, fmt.Sprintf("Error during test of package '%s': %s", packToTest, errTest.Error()))
			testErrors = append(testErrors, errTest)
		}
	}

	return testErrors
}

// BuildFolders - Runs 'go build -o <binDir>/packageName -v -ldflags <ldfFlags>' for all given packages to build
// Any package folder in the list should contain a go package with a 'go.mod' file
// - packagesToBuild: List of the packages directory path to build. Each directory should contain a 'go.mod' file
// - binDir: The output directory of the build. Any package to build will create an executable there
// - ldfFlags: Flags passed to the command via '-ldflags', may be empty
// It returns any error that may occur or nil
func BuildFolders(packagesToBuild []string, binDir, ldfFlags string) error {
	if err := EnsureDirectoryExists(binDir); err != nil {
		return err
	}

	for _, packToBuild := range packagesToBuild {
		outPutPath := filepath.Join(binDir, filepath.Base(packToBuild))
		if runtime.GOOS == "windows" {
			outPutPath = fmt.Sprintf("%s.exe", outPutPath)
		}
		fmt.Println(fmt.Sprintf("Compile package '%s' to '%s'", packToBuild, outPutPath))

		var cmd *exec.Cmd
		if ldfFlags == "" {
			fmt.Println(fmt.Sprintf("Run in %s: %s %s %s %s %s ", packToBuild, "go", "build", "-o", outPutPath, "-v"))
			cmd = exec.Command("go", "build", "-o", outPutPath, "-v")
		} else {
			fmt.Println(fmt.Sprintf("Run in %s: %s %s %s %s %s -ldflags=\"%s\"", packToBuild, "go", "build", "-o", outPutPath, "-v", ldfFlags))
			cmd = exec.Command("go", "build", "-o", outPutPath, "-v", "-ldflags", ldfFlags)
		}
		cmd.Dir = packToBuild
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		errBuild := cmd.Run()
		if errBuild != nil {
			// fmt.Println(fmt.Sprintf("Error during build of package '%s': %s", packToBuild, errBuild.Error()))
			fmt.Fprintln(os.Stderr, fmt.Sprintf("Error during build of package '%s': %s", packToBuild, errBuild.Error()))
			return errBuild
		}
	}
	return nil
}

// FindPackagesToBuild - Find a list of folders that contain go packages
// - sourceDir: The directory this function will start to search in recursively
// It returns the list of directory paths and nil in case of no error
// If an error occur the error and an empty list will be returned
func FindPackagesToBuild(sourceDir string) ([]string, error) {
	packagesToBuild := []string{}
	errFindBuild := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return nil
		}

		packToBuild := filepath.Dir(path)
		if !info.IsDir() && filepath.Base(path) == "go.mod" && !listContains(packagesToBuild, packToBuild) {
			packagesToBuild = append(packagesToBuild, packToBuild)
		}

		return nil
	})
	if errFindBuild != nil {
		return []string{}, errFindBuild
	}

	return packagesToBuild, nil
}

// FindPackagesToBuild - Find a list of folders that contain go packages with tests
// - sourceDir: The directory this function will start to search in recursively
// It returns the list of directory paths and nil in case of no error
// If an error occur the error and an empty list will be returned
func FindPackagesToTest(sourceDir string) ([]string, error) {
	packagesToTest := []string{}
	errFindTest := filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return nil
		}

		if !info.IsDir() && filepath.Ext(path) == ".go" {
			packToTest := filepath.Dir(path)
			if strings.HasSuffix(path, "_test.go") && !listContains(packagesToTest, packToTest) {
				packagesToTest = append(packagesToTest, packToTest)
			}
		}

		return nil
	})
	if errFindTest != nil {
		return []string{}, errFindTest
	}

	return packagesToTest, nil
}

// EnsureDirectoryExists - Checks if the given directory path exists, and creates it if not
// - path: The directory that should exist
// It returns any error that may occor or nil
func EnsureDirectoryExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		errCreate := os.Mkdir(path, 0755)
		if errCreate != nil {
			return errCreate
		}
	}

	return nil
}

// ReadOSDistribution - Read the linux distribution name (ID line) from '/etc/os-release'
// It returns the distribution name and nil
// In case of error it returns the error and an empty string
// Attention: This can only work on linux distributions that follow the FHS (Filesystem Hierarchy Standard)
func ReadOSDistribution() (string, error) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return "", NewOsNotSupportedByThisMethod(runtime.GOOS, "ReadOSDistribution")
	}

	ret := ""
	byteContent, err := ioutil.ReadFile("/etc/os-release")
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(byteContent), "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "ID=") {
			ret = strings.Replace(line, "ID=", "", 1)
		}
	}

	return ret, nil
}

// PathExists - check if a path exists
// - path: The file od folder path to check
// It returns true if a path exists, and false if not
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return !os.IsNotExist(err)
	}

	return true
}

func listContains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}

	return false
}

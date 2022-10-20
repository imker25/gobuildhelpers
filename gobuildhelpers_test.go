// Copyright 2022 by tobi@backfrak.de. All
// rights reserved. Use of this source code is governed
// by a BSD-style license that can be found in the
// LICENSE file.

package gobuildhelpers

import "testing"

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
}

func TestGetGitHeight(t *testing.T) {
	gitHeight, err := GetGitHeight("VersionMaster.txt", ".")

	if err != nil {
		t.Errorf("Got error '%s' but expected none", err.Error())
	}

	if gitHeight < 0 {
		t.Errorf("Got a height of '-1', but expected '0' or grater")
	}
}

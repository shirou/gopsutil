// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package host

import (
	"context"
	"testing"

	"github.com/shirou/gopsutil/v4/common"
)

func TestGetRedhatishVersion(t *testing.T) {
	var ret string
	c := []string{"Rawhide"}
	ret = getRedhatishVersion(c)
	if ret != "rawhide" {
		t.Errorf("Could not get version rawhide: %v", ret)
	}

	c = []string{"Fedora release 15 (Lovelock)"}
	ret = getRedhatishVersion(c)
	if ret != "15" {
		t.Errorf("Could not get version fedora: %v", ret)
	}

	c = []string{"Enterprise Linux Server release 5.5 (Carthage)"}
	ret = getRedhatishVersion(c)
	if ret != "5.5" {
		t.Errorf("Could not get version redhat enterprise: %v", ret)
	}

	c = []string{""}
	ret = getRedhatishVersion(c)
	if ret != "" {
		t.Errorf("Could not get version with no value: %v", ret)
	}
}

func TestGetRedhatishPlatform(t *testing.T) {
	var ret string
	c := []string{"red hat"}
	ret = getRedhatishPlatform(c)
	if ret != "redhat" {
		t.Errorf("Could not get platform redhat: %v", ret)
	}

	c = []string{"Fedora release 15 (Lovelock)"}
	ret = getRedhatishPlatform(c)
	if ret != "fedora" {
		t.Errorf("Could not get platform fedora: %v", ret)
	}

	c = []string{"Enterprise Linux Server release 5.5 (Carthage)"}
	ret = getRedhatishPlatform(c)
	if ret != "enterprise" {
		t.Errorf("Could not get platform redhat enterprise: %v", ret)
	}

	c = []string{""}
	ret = getRedhatishPlatform(c)
	if ret != "" {
		t.Errorf("Could not get platform with no value: %v", ret)
	}
}

func TestGetlsbStruct(t *testing.T) {
	cases := []struct {
		root        string
		id          string
		release     string
		codename    string
		description string
	}{
		{"arch", "Arch", "rolling", "", "Arch Linux"},
		{"ubuntu_22_04", "Ubuntu", "22.04", "jammy", "Ubuntu 22.04.2 LTS"},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.root, func(t *testing.T) {
			ctx := context.WithValue(context.Background(),
				common.EnvKey,
				common.EnvMap{common.HostEtcEnvKey: "./testdata/linux/lsbStruct/" + tt.root},
			)

			v, err := getlsbStruct(ctx)
			if err != nil {
				t.Errorf("error %v", err)
			}
			if v.ID != tt.id {
				t.Errorf("ID: want %v, got %v", tt.id, v.ID)
			}
			if v.Release != tt.release {
				t.Errorf("Release: want %v, got %v", tt.release, v.Release)
			}
			if v.Codename != tt.codename {
				t.Errorf("Codename: want %v, got %v", tt.codename, v.Codename)
			}
			if v.Description != tt.description {
				t.Errorf("Description: want %v, got %v", tt.description, v.Description)
			}

			t.Log(v)
		})
	}
}

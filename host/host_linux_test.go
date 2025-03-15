// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package host

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shirou/gopsutil/v4/common"
)

func TestGetRedhatishVersion(t *testing.T) {
	var ret string
	c := []string{"Rawhide"}
	ret = getRedhatishVersion(c)
	assert.Equalf(t, "rawhide", ret, "Could not get version rawhide: %v", ret)

	c = []string{"Fedora release 15 (Lovelock)"}
	ret = getRedhatishVersion(c)
	assert.Equalf(t, "15", ret, "Could not get version fedora: %v", ret)

	c = []string{"Enterprise Linux Server release 5.5 (Carthage)"}
	ret = getRedhatishVersion(c)
	assert.Equalf(t, "5.5", ret, "Could not get version redhat enterprise: %v", ret)

	c = []string{""}
	ret = getRedhatishVersion(c)
	assert.Emptyf(t, ret, "Could not get version with no value: %v", ret)
}

func TestGetRedhatishPlatform(t *testing.T) {
	var ret string
	c := []string{"red hat"}
	ret = getRedhatishPlatform(c)
	assert.Equalf(t, "redhat", ret, "Could not get platform redhat: %v", ret)

	c = []string{"Fedora release 15 (Lovelock)"}
	ret = getRedhatishPlatform(c)
	assert.Equalf(t, "fedora", ret, "Could not get platform fedora: %v", ret)

	c = []string{"Enterprise Linux Server release 5.5 (Carthage)"}
	ret = getRedhatishPlatform(c)
	assert.Equalf(t, "enterprise", ret, "Could not get platform redhat enterprise: %v", ret)

	c = []string{""}
	ret = getRedhatishPlatform(c)
	assert.Emptyf(t, ret, "Could not get platform with no value: %v", ret)
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
			require.NoError(t, err)
			assert.Equalf(t, v.ID, tt.id, "ID: want %v, got %v", tt.id, v.ID)
			assert.Equalf(t, v.Release, tt.release, "Release: want %v, got %v", tt.release, v.Release)
			assert.Equalf(t, v.Codename, tt.codename, "Codename: want %v, got %v", tt.codename, v.Codename)
			assert.Equalf(t, v.Description, tt.description, "Description: want %v, got %v", tt.description, v.Description)

			t.Log(v)
		})
	}
}

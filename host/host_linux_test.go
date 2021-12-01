// +build linux

package host

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestGetLSB(t *testing.T) {
	orig := os.Getenv("HOST_ETC")
	os.Setenv("HOST_ETC", "testdata/linux/ubuntu/etc")
	defer os.Setenv("HOST_ETC", orig)
	origPath := os.Getenv("PATH")
	os.Setenv("PATH", "/nonexistent")
	defer os.Setenv("PATH", origPath)

	lsb, err := getLSB()
	if err != nil {
		t.Errorf("Could not get LSB struct")
	}
	expected := &LSB{
		ID:          "Ubuntu",
		Release:     "20.04",
		Codename:    "focal",
		Description: "Ubuntu 20.04.3 LTS",
	}
	assert.Equal(t, expected, lsb)
}

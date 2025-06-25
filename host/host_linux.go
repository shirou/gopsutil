// SPDX-License-Identifier: BSD-3-Clause
//go:build linux

package host

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"golang.org/x/sys/unix"

	"github.com/shirou/gopsutil/v4/internal/common"
)

type lsbStruct struct {
	ID          string
	Release     string
	Codename    string
	Description string
}

// from utmp.h
const (
	user_PROCESS = 7 //nolint:revive //FIXME
)

func HostIDWithContext(ctx context.Context) (string, error) {
	sysProductUUID := common.HostSysWithContext(ctx, "class/dmi/id/product_uuid")
	machineID := common.HostEtcWithContext(ctx, "machine-id")
	procSysKernelRandomBootID := common.HostProcWithContext(ctx, "sys/kernel/random/boot_id")
	switch {
	// In order to read this file, needs to be supported by kernel/arch and run as root
	// so having fallback is important
	case common.PathExists(sysProductUUID):
		lines, err := common.ReadLines(sysProductUUID)
		if err == nil && len(lines) > 0 && lines[0] != "" {
			return strings.ToLower(lines[0]), nil
		}
		fallthrough
	// Fallback on GNU Linux systems with systemd, readable by everyone
	case common.PathExists(machineID):
		lines, err := common.ReadLines(machineID)
		if err == nil && len(lines) > 0 && len(lines[0]) == 32 {
			st := lines[0]
			return fmt.Sprintf("%s-%s-%s-%s-%s", st[0:8], st[8:12], st[12:16], st[16:20], st[20:32]), nil
		}
		fallthrough
	// Not stable between reboot, but better than nothing
	default:
		lines, err := common.ReadLines(procSysKernelRandomBootID)
		if err == nil && len(lines) > 0 && lines[0] != "" {
			return strings.ToLower(lines[0]), nil
		}
	}

	return "", nil
}

func numProcs(ctx context.Context) (uint64, error) {
	return common.NumProcsWithContext(ctx)
}

func BootTimeWithContext(ctx context.Context) (uint64, error) {
	return common.BootTimeWithContext(ctx, enableBootTimeCache)
}

func UptimeWithContext(_ context.Context) (uint64, error) {
	sysinfo := &unix.Sysinfo_t{}
	if err := unix.Sysinfo(sysinfo); err != nil {
		return 0, err
	}
	return uint64(sysinfo.Uptime), nil
}

func UsersWithContext(ctx context.Context) ([]UserStat, error) {
	utmpfile := common.HostVarWithContext(ctx, "run/utmp")

	file, err := os.Open(utmpfile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	count := len(buf) / sizeOfUtmp

	ret := make([]UserStat, 0, count)

	for i := 0; i < count; i++ {
		b := buf[i*sizeOfUtmp : (i+1)*sizeOfUtmp]

		var u utmp
		br := bytes.NewReader(b)
		err := binary.Read(br, binary.LittleEndian, &u)
		if err != nil {
			continue
		}
		if u.Type != user_PROCESS {
			continue
		}
		user := UserStat{
			User:     common.IntToString(u.User[:]),
			Terminal: common.IntToString(u.Line[:]),
			Host:     common.IntToString(u.Host[:]),
			Started:  int(u.Tv.Sec),
		}
		ret = append(ret, user)
	}

	return ret, nil
}

func getlsbStruct(ctx context.Context) (*lsbStruct, error) {
	ret := &lsbStruct{}
	if common.PathExists(common.HostEtcWithContext(ctx, "lsb-release")) {
		contents, err := common.ReadLines(common.HostEtcWithContext(ctx, "lsb-release"))
		if err != nil {
			return ret, err // return empty
		}
		for _, line := range contents {
			field := strings.Split(line, "=")
			if len(field) < 2 {
				continue
			}
			switch field[0] {
			case "DISTRIB_ID":
				ret.ID = strings.ReplaceAll(field[1], `"`, ``)
			case "DISTRIB_RELEASE":
				ret.Release = strings.ReplaceAll(field[1], `"`, ``)
			case "DISTRIB_CODENAME":
				ret.Codename = strings.ReplaceAll(field[1], `"`, ``)
			case "DISTRIB_DESCRIPTION":
				ret.Description = strings.ReplaceAll(field[1], `"`, ``)
			}
		}
	} else if common.PathExists("/usr/bin/lsb_release") {
		out, err := invoke.Command("/usr/bin/lsb_release")
		if err != nil {
			return ret, err
		}
		for _, line := range strings.Split(string(out), "\n") {
			field := strings.Split(line, ":")
			if len(field) < 2 {
				continue
			}
			switch field[0] {
			case "Distributor ID":
				ret.ID = strings.ReplaceAll(field[1], `"`, ``)
			case "Release":
				ret.Release = strings.ReplaceAll(field[1], `"`, ``)
			case "Codename":
				ret.Codename = strings.ReplaceAll(field[1], `"`, ``)
			case "Description":
				ret.Description = strings.ReplaceAll(field[1], `"`, ``)
			}
		}

	}

	return ret, nil
}

func PlatformInformationWithContext(ctx context.Context) (platform, family, version string, err error) {
	lsb, err := getlsbStruct(ctx)
	if err != nil {
		lsb = &lsbStruct{}
	}

	switch {
	case common.PathExistsWithContents(common.HostEtcWithContext(ctx, "oracle-release")):
		platform = "oracle"
		contents, err := common.ReadLines(common.HostEtcWithContext(ctx, "oracle-release"))
		if err == nil {
			version = getRedhatishVersion(contents)
		}

	case common.PathExistsWithContents(common.HostEtcWithContext(ctx, "enterprise-release")):
		platform = "oracle"
		contents, err := common.ReadLines(common.HostEtcWithContext(ctx, "enterprise-release"))
		if err == nil {
			version = getRedhatishVersion(contents)
		}
	case common.PathExistsWithContents(common.HostEtcWithContext(ctx, "slackware-version")):
		platform = "slackware"
		contents, err := common.ReadLines(common.HostEtcWithContext(ctx, "slackware-version"))
		if err == nil {
			version = getSlackwareVersion(contents)
		}
	case common.PathExistsWithContents(common.HostEtcWithContext(ctx, "debian_version")):
		switch lsb.ID {
		case "Ubuntu":
			platform = "ubuntu"
			version = lsb.Release
		case "LinuxMint":
			platform = "linuxmint"
			version = lsb.Release
		case "Kylin":
			platform = "Kylin"
			version = lsb.Release
		case `"Cumulus Linux"`:
			platform = "cumuluslinux"
			version = lsb.Release
		case "uos":
			platform = "uos"
			version = lsb.Release
		case "Deepin":
			platform = "Deepin"
			version = lsb.Release
		default:
			if common.PathExistsWithContents("/usr/bin/raspi-config") {
				platform = "raspbian"
			} else {
				platform = "debian"
			}
			contents, err := common.ReadLines(common.HostEtcWithContext(ctx, "debian_version"))
			if err == nil && len(contents) > 0 && contents[0] != "" {
				version = contents[0]
			}
		}
	case common.PathExistsWithContents(common.HostEtcWithContext(ctx, "neokylin-release")):
		contents, err := common.ReadLines(common.HostEtcWithContext(ctx, "neokylin-release"))
		if err == nil {
			version = getRedhatishVersion(contents)
			platform = getRedhatishPlatform(contents)
		}
	case common.PathExistsWithContents(common.HostEtcWithContext(ctx, "redhat-release")):
		contents, err := common.ReadLines(common.HostEtcWithContext(ctx, "redhat-release"))
		if err == nil {
			version = getRedhatishVersion(contents)
			platform = getRedhatishPlatform(contents)
		}
	case common.PathExistsWithContents(common.HostEtcWithContext(ctx, "system-release")):
		contents, err := common.ReadLines(common.HostEtcWithContext(ctx, "system-release"))
		if err == nil {
			version = getRedhatishVersion(contents)
			platform = getRedhatishPlatform(contents)
		}
	case common.PathExistsWithContents(common.HostEtcWithContext(ctx, "gentoo-release")):
		platform = "gentoo"
		contents, err := common.ReadLines(common.HostEtcWithContext(ctx, "gentoo-release"))
		if err == nil {
			version = getRedhatishVersion(contents)
		}
	case common.PathExistsWithContents(common.HostEtcWithContext(ctx, "SuSE-release")):
		contents, err := common.ReadLines(common.HostEtcWithContext(ctx, "SuSE-release"))
		if err == nil {
			version = getSuseVersion(contents)
			platform = getSusePlatform(contents)
		}
		// TODO: slackware detection
	case common.PathExistsWithContents(common.HostEtcWithContext(ctx, "arch-release")):
		platform = "arch"
		version = lsb.Release
	case common.PathExistsWithContents(common.HostEtcWithContext(ctx, "alpine-release")):
		platform = "alpine"
		contents, err := common.ReadLines(common.HostEtcWithContext(ctx, "alpine-release"))
		if err == nil && len(contents) > 0 && contents[0] != "" {
			version = contents[0]
		}
	case common.PathExistsWithContents(common.HostEtcWithContext(ctx, "os-release")):
		p, v, err := common.GetOSReleaseWithContext(ctx)
		if err == nil {
			platform = p
			version = v
		}
	case lsb.ID == "RedHat":
		platform = "redhat"
		version = lsb.Release
	case lsb.ID == "Amazon":
		platform = "amazon"
		version = lsb.Release
	case lsb.ID == "ScientificSL":
		platform = "scientific"
		version = lsb.Release
	case lsb.ID == "XenServer":
		platform = "xenserver"
		version = lsb.Release
	case lsb.ID != "":
		platform = strings.ToLower(lsb.ID)
		version = lsb.Release
	}

	platform = strings.Trim(platform, `"`)

	switch platform {
	case "debian", "ubuntu", "linuxmint", "raspbian", "Kylin", "cumuluslinux", "uos", "Deepin":
		family = "debian"
	case "fedora":
		family = "fedora"
	case "oracle", "centos", "redhat", "scientific", "enterpriseenterprise", "amazon", "xenserver", "cloudlinux", "ibm_powerkvm", "rocky", "almalinux":
		family = "rhel"
	case "suse", "opensuse", "opensuse-leap", "opensuse-tumbleweed", "opensuse-tumbleweed-kubic", "sles", "sled", "caasp":
		family = "suse"
	case "gentoo":
		family = "gentoo"
	case "slackware":
		family = "slackware"
	case "arch":
		family = "arch"
	case "exherbo":
		family = "exherbo"
	case "alpine":
		family = "alpine"
	case "coreos":
		family = "coreos"
	case "solus":
		family = "solus"
	case "neokylin":
		family = "neokylin"
	case "anolis":
		family = "anolis"
	}

	return platform, family, version, nil
}

func KernelVersionWithContext(_ context.Context) (version string, err error) {
	var utsname unix.Utsname
	err = unix.Uname(&utsname)
	if err != nil {
		return "", err
	}
	return unix.ByteSliceToString(utsname.Release[:]), nil
}

func getSlackwareVersion(contents []string) string {
	c := strings.ToLower(strings.Join(contents, ""))
	c = strings.Replace(c, "slackware ", "", 1)
	return c
}

var redhatishReleaseMatch = regexp.MustCompile(`release (\w[\d.]*)`)

func getRedhatishVersion(contents []string) string {
	c := strings.ToLower(strings.Join(contents, ""))

	if strings.Contains(c, "rawhide") {
		return "rawhide"
	}
	if matches := redhatishReleaseMatch.FindStringSubmatch(c); matches != nil {
		return matches[1]
	}
	return ""
}

func getRedhatishPlatform(contents []string) string {
	c := strings.ToLower(strings.Join(contents, ""))

	if strings.Contains(c, "red hat") {
		return "redhat"
	}
	f := strings.Split(c, " ")

	return f[0]
}

var (
	suseVersionMatch    = regexp.MustCompile(`VERSION = ([\d.]+)`)
	susePatchLevelMatch = regexp.MustCompile(`PATCHLEVEL = (\d+)`)
)

func getSuseVersion(contents []string) string {
	version := ""
	for _, line := range contents {
		if matches := suseVersionMatch.FindStringSubmatch(line); matches != nil {
			version = matches[1]
		} else if matches = susePatchLevelMatch.FindStringSubmatch(line); matches != nil {
			version = version + "." + matches[1]
		}
	}
	return version
}

func getSusePlatform(contents []string) string {
	c := strings.ToLower(strings.Join(contents, ""))
	if strings.Contains(c, "opensuse") {
		return "opensuse"
	}
	return "suse"
}

func VirtualizationWithContext(ctx context.Context) (string, string, error) {
	return common.VirtualizationWithContext(ctx)
}

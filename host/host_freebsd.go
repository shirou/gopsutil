//go:build freebsd
// +build freebsd

package host

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unsafe"

	"github.com/shirou/gopsutil/v3/internal/common"
	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sys/unix"
)

const (
	UTNameSize = 16 /* see MAXLOGNAME in <sys/param.h> */
	UTLineSize = 8
	UTHostSize = 16
)

func HostIDWithContext(ctx context.Context) (string, error) {
	uuid, err := unix.Sysctl("kern.hostuuid")
	if err != nil {
		return "", err
	}
	return strings.ToLower(uuid), err
}

func numProcs(ctx context.Context) (uint64, error) {
	procs, err := process.PidsWithContext(ctx)
	if err != nil {
		return 0, err
	}
	return uint64(len(procs)), nil
}

func UsersWithContext(ctx context.Context) ([]UserStat, error) {
	utmpfile := "/var/run/utx.active"
	if !common.PathExists(utmpfile) {
		utmpfile = "/var/run/utmp" // before 9.0
		return getUsersFromUtmp(utmpfile)
	}

	var ret []UserStat
	file, err := os.Open(utmpfile)
	if err != nil {
		return ret, err
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return ret, err
	}

	entrySize := sizeOfUtmpx
	count := len(buf) / entrySize

	for i := 0; i < count; i++ {
		b := buf[i*sizeOfUtmpx : (i+1)*sizeOfUtmpx]
		var u Utmpx
		br := bytes.NewReader(b)
		err := binary.Read(br, binary.BigEndian, &u)
		if err != nil || u.Type != 4 {
			continue
		}
		sec := math.Floor(float64(u.Tv) / 1000000)
		user := UserStat{
			User:     common.IntToString(u.User[:]),
			Terminal: common.IntToString(u.Line[:]),
			Host:     common.IntToString(u.Host[:]),
			Started:  int(sec),
		}

		ret = append(ret, user)
	}

	return ret, nil
}

func PlatformInformationWithContext(ctx context.Context) (string, string, string, error) {
	platform, err := unix.Sysctl("kern.ostype")
	if err != nil {
		return "", "", "", err
	}

	version, err := unix.Sysctl("kern.osrelease")
	if err != nil {
		return "", "", "", err
	}

	return strings.ToLower(platform), "", strings.ToLower(version), nil
}

func VirtualizationWithContext(ctx context.Context) (string, string, error) {
	return "", "", common.ErrNotImplementedError
}

// before 9.0
func getUsersFromUtmp(utmpfile string) ([]UserStat, error) {
	var ret []UserStat
	file, err := os.Open(utmpfile)
	if err != nil {
		return ret, err
	}
	defer file.Close()

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return ret, err
	}

	u := Utmp{}
	entrySize := int(unsafe.Sizeof(u))
	count := len(buf) / entrySize

	for i := 0; i < count; i++ {
		b := buf[i*entrySize : i*entrySize+entrySize]
		var u Utmp
		br := bytes.NewReader(b)
		err := binary.Read(br, binary.LittleEndian, &u)
		if err != nil || u.Time == 0 {
			continue
		}
		user := UserStat{
			User:     common.IntToString(u.Name[:]),
			Terminal: common.IntToString(u.Line[:]),
			Host:     common.IntToString(u.Host[:]),
			Started:  int(u.Time),
		}

		ret = append(ret, user)
	}

	return ret, nil
}

func SensorsTemperaturesWithContext(ctx context.Context) ([]TemperatureStat, error) {
	sysctl, err := exec.LookPath("sysctl")
	var ret []TemperatureStat
	if err != nil {
		return ret, err
	}

	out, err := invoke.CommandWithContext(context.WithValue(ctx, common.InvokerCtxKeyEnv, []string{"LC_ALL=C"}), sysctl, "-a")
	if err != nil {
		return ret, err
	}

	tjmaxs := make(map[string]float64)

	var warns Warnings

	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		// hw.acpi.thermal.tz1.temperature: 29.9C
		// dev.cpu.7.temperature: 28.0C
		// dev.cpu.7.coretemp.tjmax: 100.0C
		flds := strings.SplitN(sc.Text(), ":", 2)
		if len(flds) != 2 ||
			!(strings.HasSuffix(flds[0], ".temperature") ||
				strings.HasSuffix(flds[0], ".coretemp.tjmax")) ||
			!strings.HasSuffix(flds[1], "C") {
			continue
		}
		v, err := strconv.ParseFloat(strings.TrimSuffix(strings.TrimSpace(flds[1]), "C"), 64)
		if err != nil {
			warns.Add(err)
			continue
		}
		k := strings.TrimSuffix(flds[0], ".temperature")
		if k == flds[0] {
			k = strings.TrimSuffix(flds[0], ".coretemp.tjmax")
			tjmaxs[k] = v
			continue
		}
		ts := TemperatureStat{
			SensorKey:   k,
			Temperature: v,
		}
		ret = append(ret, ts)
	}

	for i, ts := range ret[:] {
		if tjmax, ok := tjmaxs[ts.SensorKey]; ok {
			ts.Critical = tjmax
			ret[i] = ts
		}
	}

	return ret, warns.Reference()
}

func KernelVersionWithContext(ctx context.Context) (string, error) {
	_, _, version, err := PlatformInformationWithContext(ctx)
	return version, err
}

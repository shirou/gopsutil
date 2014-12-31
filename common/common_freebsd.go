// +build freebsd

package common

import (
	"os/exec"
	"strings"
)

func DoSysctrl(mib string) ([]string, error) {
	out, err := exec.Command("/sbin/sysctl", "-n", mib).Output()
	if err != nil {
		return []string{}, err
	}
	v := strings.Replace(string(out), "{ ", "", 1)
	v = strings.Replace(string(v), " }", "", 1)
	values := strings.Fields(string(v))

	return values, nil
}

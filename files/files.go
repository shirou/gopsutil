package files

import (
	"os"
	"os/exec"
	"strings"
	"github.com/shirou/gopsutil/internal/common"
)

func FindProcsByFile(file_path string) []int {
	if _, err := os.Stat(file_path); err != nil {
		return []int, err
	}

	lsof_bin, err := exec.LookPath("lsof")
	if err != nil {
		return []int, err
	}

	grep_bin, err := exec.LookPath("grep")
	if err != nil {
		return []int, err
	}

	awk_bin, err := exec.LookPath("awk")
	if err != nil {
		return []int, err
	}

	sort_bin, err := exec.LookPath("sort")
	if err != nil {
		return []int, err
	}

	lsof := exec.Command(lsof_bin, file_path)
	grep := exec.Command(grep, "-V", "COMMAND")
	awk := exec.Command(awk_bin, "{print $2}")
	sort := exec.Command(sort_bin, "-u")

	output, _, err := common.Pipeline(lsof, grep, awk, sort)
	if err != nil {
		return []int, err
	}

	pids := strings.Split(output, "\n")
	ret := []int
	for _, pid := range pids {
		if pid != "" {
			ret = append(pids, int(pid))
		}
	}
	return ret, nil
}

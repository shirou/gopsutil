package files

import (
	"os"
	"os/exec"
	"strings"
	"strconv"
	"github.com/shirou/gopsutil/internal/common"
)

func FindProcsByFile(file_path string) ([]int, error) {
	if _, err := os.Stat(file_path); err != nil {
		return []int{}, err
	}

	lsof_bin, err := exec.LookPath("lsof")
	if err != nil {
		return []int{}, err
	}

	awk_bin, err := exec.LookPath("awk")
	if err != nil {
		return []int{}, err
	}

	sort_bin, err := exec.LookPath("sort")
	if err != nil {
		return []int{}, err
	}

	lsof := exec.Command(lsof_bin, file_path)
	awk := exec.Command(awk_bin, "NR>1 {print $2}")
	sort := exec.Command(sort_bin, "-u")

	output, _, err := common.Pipeline(lsof, awk, sort)
	if err != nil {
		return []int{}, err
	}

	pids := strings.Split(string(output), "\n")
	ret := []int{}
	for _, pid := range pids {
		if pid != "" {
			int_pid, _ := strconv.Atoi(pid)
			ret = append(ret, int_pid)
		}
	}
	return ret, nil
}

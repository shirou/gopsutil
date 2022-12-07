package setget

import "github.com/shirou/gopsutil/v3/internal/common"

func SetPathPrefix(prefix string) {
	common.PathPrefix = prefix
}

func GetPathPrefix() string {
	return common.PathPrefix
}

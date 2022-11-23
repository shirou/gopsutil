package packages

import (
	"errors"
)

type log struct {
	User      string   `json:"user"`
	Date      string   `json:"date"`
	Command   string   `json:"command"`
	Installed []string `json:"installed"`
	Upgraded  []string `json:"upgraded"`
	Purged    []string `json:"purged"`
}

type pkgManagerInfo struct {
	InstalledPackages []string `json:"installedpackages"`
	UpgradedPackages  []string `json:"upgradedpackages"`
	PurgedPackages    []string `json:"purgedpackages"`
}

type pkgManagerLogs struct {
	Logs []log
}

var supportedPackageManagers = map[string]func() (pkgManagerLogs, error){
	"apt": aptParser,
	"dnf": dnfParser,
}

func PackageManagerLogs(pktMngr string) (pkgManagerLogs, error) {
	if _, av := supportedPackageManagers[pktMngr]; !av {
		return pkgManagerLogs{}, errors.New("package manager is not supported")
	}
	return supportedPackageManagers[pktMngr]()

}

func PackageManagerInfo(pktMngr string) (pkgManagerInfo, error) {

	aggLogs, err := supportedPackageManagers[pktMngr]()

	if err != nil {
		return pkgManagerInfo{}, err
	}
	rtrnValue := pkgManagerInfo{}

	for _, v := range aggLogs.Logs {

		rtrnValue.InstalledPackages = append(rtrnValue.InstalledPackages, v.Installed...)
		rtrnValue.UpgradedPackages = append(rtrnValue.UpgradedPackages, v.Upgraded...)
		rtrnValue.PurgedPackages = append(rtrnValue.PurgedPackages, v.Purged...)

	}
	return rtrnValue, nil
}

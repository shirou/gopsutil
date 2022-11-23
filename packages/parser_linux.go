package packages

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var aptRegex, _ = regexp.Compile(`(?:[^,(]|\([^)]*\))+`)

func aptParser() (mngInfo pkgManagerLogs, err error) {

	f, err := os.Open("/var/log/apt/history.log")
	if err != nil {
		return
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)

	logs := [][]string{}

startPoint:
	logLine := []string{}
	for n := 0; s.Scan(); n++ {
		// skip header line
		if n == 0 {
			continue
		}
		// empty line signifies new log
		if s.Text() == "" {
			logs = append(logs, logLine)
			goto startPoint
		}
		logLine = append(logLine, s.Text())

	}
	logs = append(logs, logLine)

	for _, v := range logs {
		parsed, err := aptLineParse(v)
		if err != nil {
			continue
		}
		mngInfo.Logs = append(mngInfo.Logs, parsed)
	}

	return
}

func aptLineParse(logLines []string) (log, error) {

	rtrnLog := log{}

	for _, line := range logLines {
		idx := strings.SplitN(line, ":", 2)
		if len(idx) < 2 {
			return rtrnLog, errors.New("not valid rawline for apt log")
		}

		if idx[0] == "Start-Date" {
			rtrnLog.Date = fmt.Sprintf("%q", idx[1])
		} else if idx[0] == "Commandline" {
			rtrnLog.Command = fmt.Sprintf("%q", idx[1])
		} else if idx[0] == "Requested-By" {
			rtrnLog.User = fmt.Sprintf("%q", idx[1])
		} else if idx[0] == "Install" {
			pkgs := aptRegex.FindAll([]byte(idx[1]), -1)
			for _, v := range pkgs {
				rtrnLog.Installed = append(rtrnLog.Installed, fmt.Sprintf("%q", v))
			}
		} else if idx[0] == "Upgrade" {
			pkgs := aptRegex.FindAll([]byte(idx[1]), -1)
			for _, v := range pkgs {
				rtrnLog.Upgraded = append(rtrnLog.Upgraded, fmt.Sprintf("%q", v))
			}
		} else if idx[0] == "Purge" {
			pkgs := aptRegex.FindAll([]byte(idx[1]), -1)
			for _, v := range pkgs {
				rtrnLog.Purged = append(rtrnLog.Purged, fmt.Sprintf("%q", v))
			}
		}

	}
	return rtrnLog, nil
}

func dnfParser() (mngInfo pkgManagerLogs, err error) {
	f, err := os.Open("/var/log/dnf.rpm.log")
	if err != nil {
		return pkgManagerLogs{}, err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)

	for s.Scan() {
		logData := log{}
		logLine := strings.Split(s.Text(), "SUBDEBUG")
		if len(logLine) < 2 {
			continue
		}
		data := strings.Split(logLine[1], ":")

		logData.Date = logLine[0]

		if strings.TrimSpace(data[0]) == "Installed" {
			logData.Installed = append(logData.Installed, data[1])
		} else if strings.TrimSpace(data[0]) == "Upgraded" {
			logData.Upgraded = append(logData.Upgraded, data[1])
		} else if strings.TrimSpace(data[0]) == "Erase" {
			logData.Purged = append(logData.Purged, data[1])
		}

		mngInfo.Logs = append(mngInfo.Logs, logData)
	}
	return
}

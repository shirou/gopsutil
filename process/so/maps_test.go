// +build linux

package so

import (
	"bufio"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLibsFromMaps(t *testing.T) {
	r := strings.NewReader(mapsFile)
	libs := parseMaps(bufio.NewReader(r), AllLibraries)
	expected := []string{
		"/usr/lib/x86_64-linux-gnu/libc-2.31.so",
		"/usr/lib/x86_64-linux-gnu/ld-2.31.so",
	}
	assert.ElementsMatch(t, expected, libs)
}

func TestLibsFromMapsWithFilter(t *testing.T) {
	filter := regexp.MustCompile("libc")
	r := strings.NewReader(mapsFile)
	libs := parseMaps(bufio.NewReader(r), filter)
	expected := []string{"/usr/lib/x86_64-linux-gnu/libc-2.31.so"}
	assert.ElementsMatch(t, expected, libs)
}

var mapsFile = `
7f178d0a6000-7f178d0cb000 r--p 00000000 fd:00 268741                     /usr/lib/x86_64-linux-gnu/libc-2.31.so
7f178d0cb000-7f178d243000 r-xp 00025000 fd:00 268741                     /usr/lib/x86_64-linux-gnu/libc-2.31.so
7f178d243000-7f178d28d000 r--p 0019d000 fd:00 268741                     /usr/lib/x86_64-linux-gnu/libc-2.31.so
7f178d28d000-7f178d28e000 ---p 001e7000 fd:00 268741                     /usr/lib/x86_64-linux-gnu/libc-2.31.so
7f178d28e000-7f178d291000 r--p 001e7000 fd:00 268741                     /usr/lib/x86_64-linux-gnu/libc-2.31.so
7f178d291000-7f178d294000 rw-p 001ea000 fd:00 268741                     /usr/lib/x86_64-linux-gnu/libc-2.31.so
7f178d294000-7f178d29a000 rw-p 00000000 00:00 0
7f178d29a000-7f178d29b000 r--p 00000000 fd:00 262340                     /usr/lib/locale/C.UTF-8/LC_TELEPHONE
7f178d29b000-7f178d29c000 r--p 00000000 fd:00 262333                     /usr/lib/locale/C.UTF-8/LC_MEASUREMENT
7f178d29c000-7f178d2a3000 r--s 00000000 fd:00 269008                     /usr/lib/x86_64-linux-gnu/gconv/gconv-modules.cache
7f178d2a3000-7f178d2a4000 r--p 00000000 fd:00 268737                     /usr/lib/x86_64-linux-gnu/ld-2.31.so
7f178d2a4000-7f178d2c7000 r-xp 00001000 fd:00 268737                     /usr/lib/x86_64-linux-gnu/ld-2.31.so
7f178d2c7000-7f178d2cf000 r--p 00024000 fd:00 268737                     /usr/lib/x86_64-linux-gnu/ld-2.31.so
7f178d2cf000-7f178d2d0000 r--p 00000000 fd:00 262317                     /usr/lib/locale/C.UTF-8/LC_IDENTIFICATION
7f178d2d0000-7f178d2d1000 r--p 0002c000 fd:00 268737                     /usr/lib/x86_64-linux-gnu/ld-2.31.so
7f178d2d1000-7f178d2d2000 rw-p 0002d000 fd:00 268737                     /usr/lib/x86_64-linux-gnu/ld-2.31.so
7f178d2d2000-7f178d2d3000 rw-p 00000000 00:00 0
7ffe712a4000-7ffe712c5000 rw-p 00000000 00:00 0                          [stack]
7ffe71317000-7ffe7131a000 r--p 00000000 00:00 0                          [vvar]
7ffe7131a000-7ffe7131b000 r-xp 00000000 00:00 0                          [vdso]
ffffffffff600000-ffffffffff601000 --xp 00000000 00:00 0                  [vsyscall]
`

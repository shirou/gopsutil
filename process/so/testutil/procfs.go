// +build linux

package testutil

/*
   Here the process included in procfs directory with the help of update-procfs.sh
   Libraries are extracted from procfs with the oneliner : awk '{print $6}' procfs/138039/maps | sort -u | grep -v '^\[' | xargs -n 1 --replace=ABC echo '"ABC",'

      1 ?        Ss     0:04 /sbin/init
   2211 pts/2    Ss     0:02  \_ /bin/bash
 138104 pts/2    T      0:00      \_ sudo runc exec sleep999 sleep 999
 138105 pts/2    Rl     5:19      |   \_ runc exec sleep999 sleep 999
 138112 pts/2    T      0:00      |       \_ runc exec sleep999 sleep 999
 370680 pts/2    S      0:00      \_ sleep 99

$ sudo runc list
ID          PID         STATUS      BUNDLE                                                                            CREATED                          OWNER
sleep999    138039      created     /FAKEHOME/devel/go/src/github.com/DataDog/datadog-agent/pkg/network/so/cont   2021-08-11T11:18:42.458045668Z   root

*/

// Process define one process that is defined in procfs directory
type Process struct {
	Pid       int
	Cmdline   string
	Libraries []string
}

// ProcFS is the image of the process present in procfs directory
var ProcFS = []Process{
	{
		Pid:     1,
		Cmdline: "/sbin/init",
		Libraries: []string{
			"/usr/lib/ld-2.33.so",
			"/usr/lib/libacl.so.1.1.2301",
			"/usr/lib/libaudit.so.1.0.0",
			"/usr/lib/libblkid.so.1.1.0",
			"/usr/lib/libc-2.33.so",
			"/usr/lib/libcap-ng.so.0.0.0",
			"/usr/lib/libcap.so.2.51",
			"/usr/lib/libcrypto.so.1.1",
			"/usr/lib/libcrypt.so.2.0.0",
			"/usr/lib/libdl-2.33.so",
			"/usr/lib/libffi.so.7.1.0",
			"/usr/lib/libgcc_s.so.1",
			"/usr/lib/libgcrypt.so.20.3.3",
			"/usr/lib/libgpg-error.so.0.32.0",
			"/usr/lib/libip4tc.so.2.0.0",
			"/usr/lib/libkmod.so.2.3.7",
			"/usr/lib/liblz4.so.1.9.3",
			"/usr/lib/liblzma.so.5.2.5",
			"/usr/lib/libmount.so.1.1.0",
			"/usr/lib/libp11-kit.so.0.3.0",
			"/usr/lib/libpam.so.0.85.1",
			"/usr/lib/libpthread-2.33.so",
			"/usr/lib/librt-2.33.so",
			"/usr/lib/libseccomp.so.2.5.1",
			"/usr/lib/libz.so.1.2.11",
			"/usr/lib/libzstd.so.1.5.0",
			"/usr/lib/systemd/libsystemd-shared-248.so",
			"/usr/lib/systemd/systemd",
		},
	},
	{
		Pid:     2211,
		Cmdline: "/bin/bash",
		Libraries: []string{
			"/usr/bin/bash",
			"/usr/lib/ld-2.33.so",
			"/usr/lib/libc-2.33.so",
			"/usr/lib/libcap.so.2.51",
			"/usr/lib/libcrypto.so.1.1",
			"/usr/lib/libcrypt.so.2.0.0",
			"/usr/lib/libdl-2.33.so",
			"/usr/lib/libffi.so.7.1.0",
			"/usr/lib/libgcc_s.so.1",
			"/usr/lib/libncursesw.so.6.2",
			"/usr/lib/libnss_files-2.33.so",
			"/usr/lib/libnss_mymachines.so.2",
			"/usr/lib/libnss_systemd.so.2",
			"/usr/lib/libp11-kit.so.0.3.0",
			"/usr/lib/libpthread-2.33.so",
			"/usr/lib/libreadline.so.8.1",
			"/usr/lib/librt-2.33.so",
			"/usr/lib/locale/locale-archive",
			"/usr/share/locale/en_GB/LC_MESSAGES/libc.mo",
		},
	},
	{
		Pid:     138104,
		Cmdline: "sudo runc exec sleep999 sleep 999",
		Libraries: []string{
			"/usr/bin/sudo",
			"/usr/lib/ld-2.33.so",
			"/usr/lib/libaudit.so.1.0.0",
			"/usr/lib/libc-2.33.so",
			"/usr/lib/libcap-ng.so.0.0.0",
			"/usr/lib/libcap.so.2.51",
			"/usr/lib/libcom_err.so.2.1",
			"/usr/lib/libcrypto.so.1.1",
			"/usr/lib/libcrypt.so.2.0.0",
			"/usr/lib/libdl-2.33.so",
			"/usr/lib/libffi.so.7.1.0",
			"/usr/lib/libgcc_s.so.1",
			"/usr/lib/libgssapi_krb5.so.2.2",
			"/usr/lib/libk5crypto.so.3.1",
			"/usr/lib/libkeyutils.so.1.10",
			"/usr/lib/libkrb5.so.3.3",
			"/usr/lib/libkrb5support.so.0.1",
			"/usr/lib/liblber-2.4.so.2.11.7",
			"/usr/lib/libldap-2.4.so.2.11.7",
			"/usr/lib/libnss_files-2.33.so",
			"/usr/lib/libnss_mymachines.so.2",
			"/usr/lib/libnss_systemd.so.2",
			"/usr/lib/libp11-kit.so.0.3.0",
			"/usr/lib/libpam.so.0.85.1",
			"/usr/lib/libpthread-2.33.so",
			"/usr/lib/libresolv-2.33.so",
			"/usr/lib/librt-2.33.so",
			"/usr/lib/libsasl2.so.3.0.0",
			"/usr/lib/libssl.so.1.1",
			"/usr/lib/libtirpc.so.3.0.0",
			"/usr/lib/libutil-2.33.so",
			"/usr/lib/libz.so.1.2.11",
			"/usr/lib/locale/locale-archive",
			"/usr/lib/security/pam_deny.so",
			"/usr/lib/security/pam_env.so",
			"/usr/lib/security/pam_faillock.so",
			"/usr/lib/security/pam_limits.so",
			"/usr/lib/security/pam_permit.so",
			"/usr/lib/security/pam_systemd_home.so",
			"/usr/lib/security/pam_time.so",
			"/usr/lib/security/pam_unix.so",
			"/usr/lib/security/pam_warn.so",
			"/usr/lib/sudo/libsudo_util.so.0.0.0",
			"/usr/lib/sudo/sudoers.so",
		},
	},
	{
		Pid:     138105,
		Cmdline: "runc exec sleep999 sleep 999",
		Libraries: []string{
			"/usr/bin/runc",
			"/usr/lib/ld-2.33.so",
			"/usr/lib/libc-2.33.so",
			"/usr/lib/libpthread-2.33.so",
			"/usr/lib/libseccomp.so.2.5.1",
		},
	},
	{
		Pid:     138112,
		Cmdline: "runc exec sleep999 sleep 999",
		Libraries: []string{
			"/usr/bin/runc",
			"/usr/lib/ld-2.33.so",
			"/usr/lib/libc-2.33.so",
			"/usr/lib/libpthread-2.33.so",
			"/usr/lib/libseccomp.so.2.5.1",
		},
	},
	{
		Pid:     370680,
		Cmdline: "sleep 99",
		Libraries: []string{
			"/usr/bin/sleep",
			"/usr/lib/ld-2.33.so",
			"/usr/lib/libc-2.33.so",
			"/usr/lib/locale/locale-archive",
		},
	},
	{
		Pid:     138039,
		Cmdline: "runc init",
		Libraries: []string{
			"/",
			"/usr/lib/ld-2.33.so",
			"/usr/lib/libc-2.33.so",
			"/usr/lib/libpthread-2.33.so",
			"/usr/lib/libseccomp.so.2.5.1",
		},
	},
}

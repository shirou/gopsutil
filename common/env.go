package common

type EnvKeyType string

// EnvKey is a context key that can be used to set programmatically the environment
// gopsutil relies on to perform calls against the OS.
// Example of use:
//
//	ctx := context.WithValue(context.Background(), common.EnvKey, EnvMap{"HOST_PROC": "/myproc"})
//	avg, err := load.AvgWithContext(ctx)
var EnvKey = EnvKeyType("env")

const (
	HostProcEnvKey EnvKeyType = "HOST_PROC"
	HostSysEnvKey  EnvKeyType = "HOST_SYS"
	HostEtcEnvKey  EnvKeyType = "HOST_ETC"
)

type EnvMap map[EnvKeyType]string

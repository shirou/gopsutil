package common

type envKey string

// Env is a context key that can be used to set programmatically the environment
// gopsutil relies on to perform calls against the OS.
// Example of use:
//
//	ctx := context.WithValue(context.Background(), Env, map[string]string{"HOST_PROC": "/myproc"})
//	avg, err := load.AvgWithContext(ctx)
var Env = envKey("env")

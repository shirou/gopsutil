# Regenerating Linux host definitions

From the repository root:

```sh
docker build --platform linux/amd64 -t gopsutil-host-godefs -f host/linux/Dockerfile .
docker run --rm --platform linux/amd64 --user "$(id -u):$(id -g)" \
  -v "$PWD:/src" gopsutil-host-godefs
```

This regenerates all Linux `host_linux_<arch>.go` files from
`host/types_linux.go`. Append architecture names to regenerate only those files,
for example `gopsutil-host-godefs ppc64 ppc64le s390x`.

The image uses Debian trixie's available Go, GCC, and glibc packages. Their versions
are recorded in each generated file; rebuilding the image may change the output.
QEMU is not needed for cross-generation, but non-amd64 hosts need amd64 image emulation.

The script marks two postprocessing adjustments with `// changed by hand`:

- Replace `_Ctype_struct___0` with the generated `utTv` type.
- Convert `Line`, `Id`, `User`, and `Host` byte arrays to `int8` for `common.IntToString`.

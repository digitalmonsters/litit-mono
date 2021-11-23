//go:build linux
// +build linux

package profiler

import (
	"fmt"
	"os/exec"
	"path"
)

func (p *Profiler) collectStats(finalDir string) error {
	return exec.Command(
		"/bin/sh", "-c",
		fmt.Sprintf(`ps -eo pcpu,pid,user,%%mem,args | sort -r -k1 > "%s"`, path.Join(finalDir, "ps_out.txt")),
	).Run()
}

func (p *Profiler) collectIoStat(finalDir string) error {
	return exec.Command(
		"/bin/sh", "-c",
		fmt.Sprintf(`pidstat -dl > "%s"`, path.Join(finalDir, "pidstat.txt")),
	).Run()
}

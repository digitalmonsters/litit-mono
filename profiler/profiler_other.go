//go:build !linux
// +build !linux

package profiler

import "errors"

func (p *Profiler) collectStats(finalDir string) error {
	return errors.New("[Profiler] ps is not supported on current platform")
}

func (p *Profiler) collectIoStat(finalDir string) error {
	return errors.New("[Profiler] pidstat is not supported on current platform")
}

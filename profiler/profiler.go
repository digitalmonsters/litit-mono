package profiler

import (
	"bufio"
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/shirou/gopsutil/v3/cpu"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"sort"
	"time"
)

type Config struct {
	ProfilerDirectory              string `json:"profiler_directory"`
	MinPauseBetweenProfilesSeconds int    `json:"min_pause_between_profiles_seconds"`
	ProfileSessionLengthSeconds    int    `json:"profile_session_length_seconds"`
	DumpEveryMinutes               int    `json:"dump_every_minutes"`
	CpuLoadThresholdPercent        int    `json:"cpu_load_threshold_percent"`
	MaxDumpsInFolder               int    `json:"max_dumps_in_folder"`
	MemLoadThresholdMb             int    `json:"mem_load_threshold_mb"`
}

type Profiler struct {
	dumpCounter    int
	isRunning      bool
	lastProfiledAt time.Time
	baseLogger     zerolog.Logger
	config         Config
	context        context.Context
}

type RunType string

const (
	RunTypeTime = RunType("time")
	RunTypeLoad = RunType("load")
)

func NewProfiler(config Config, logger zerolog.Logger, context context.Context) *Profiler {
	pr := &Profiler{config: config, baseLogger: logger, context: context}

	return pr
}

func (p *Profiler) StartMonitoring() {
	if p.config.DumpEveryMinutes > 0 {
		p.baseLogger.Info().Msgf("[Profiler] Profiler will run every %v minutes", p.config.DumpEveryMinutes)
		go func() {
			for p.context.Err() == nil {
				time.Sleep(time.Duration(p.config.DumpEveryMinutes) * time.Minute)

				if _, err := p.makeSnapshotInternal(RunTypeTime); err != nil {
					p.baseLogger.Err(err).Send()
				}
			}
		}()
	} else {
		p.baseLogger.Info().Msg("[Profiler] dump_every_minutes is 0, so no time-based profiles")
	}

	go func() {
		p.baseLogger.Info().Msg("[Profiler] CPU Monitoring started")

		for p.context.Err() == nil {
			timings, err := cpu.PercentWithContext(p.context, 20*time.Second, false)

			if err != nil {
				p.baseLogger.Err(err).Msg("[Profiler] Can not calculate PercentWithContext")
			} else {
				p.baseLogger.Trace().Msgf("cpu timings : %v", timings)
			}

			if len(timings) == 0 {
				p.baseLogger.Err(err).Msg("[Profiler] Timings should not be empty")
				time.Sleep(10 * time.Second)
				continue
			}

			totalUsageMem := float64(0)
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			if m.TotalAlloc > 0 {
				totalUsageMem = float64(m.TotalAlloc) / 1024 / 1024

				p.baseLogger.Trace().Msgf("mem usage : %v MB", totalUsageMem)
			}

			if timings[0] >= float64(p.config.CpuLoadThresholdPercent) || totalUsageMem >= float64(p.config.MemLoadThresholdMb) {
				if _, err := p.MakeSnapshot(RunTypeLoad); err != nil {
					p.baseLogger.Err(err).Send()
				}
			}
		}
	}()
}

func (p *Profiler) MakeSnapshot(runType RunType) (string, error) {
	return p.makeSnapshotInternal(runType)
}

func (p *Profiler) cleanupFolder(folderToCheck string, logger zerolog.Logger) error {
	if p.config.MaxDumpsInFolder == 0 {
		return nil
	}

	items, err := ioutil.ReadDir(folderToCheck)

	if err != nil {
		return errors.WithStack(err)
	}

	var folders []string

	for _, folder := range items {
		if !folder.IsDir() { // only dirs are important for us
			continue
		}

		folders = append(folders, folder.Name())
	}

	if len(folders) < p.config.MaxDumpsInFolder {
		return nil
	}

	sort.Strings(folders)

	toRemove := len(folders) - p.config.MaxDumpsInFolder
	foldersToRemove := folders[:toRemove]

	for _, item := range foldersToRemove {
		toRemoveFolder := path.Join(folderToCheck, item)
		logger.Info().Msgf("[Profiler] Will remove %v", toRemoveFolder)

		if err := os.RemoveAll(toRemoveFolder); err != nil {
			logger.Err(err).Msg("[Profiler] Can not remove folder")
		}
	}

	return nil
}

func (p *Profiler) makeSnapshotInternal(runType RunType) (string, error) {
	if p.lastProfiledAt.Add(time.Duration(p.config.MinPauseBetweenProfilesSeconds) * time.Second).After(time.Now().UTC()) {
		return "", errors.New(fmt.Sprintf("[Profiller] Not allowed. Should wait at least for %v. Last profile session was %v",
			p.config.MinPauseBetweenProfilesSeconds, p.lastProfiledAt))
	}

	if !p.isRunning {
		p.isRunning = true
	} else {
		return "", errors.New("[Profiler] Session already running")
	}

	defer func() {
		p.isRunning = false
	}()

	p.isRunning = true
	currentDir, err := os.Getwd()

	if err != nil {
		return "", errors.WithStack(err)
	}

	p.dumpCounter += 1

	dirName := fmt.Sprintf("%v___%v", time.Now().UTC().Format("2006_01_02__15_04_05"), p.dumpCounter)

	baseDirectory := path.Join(currentDir, p.config.ProfilerDirectory, string(runType))

	finalFolder := path.Join(baseDirectory, dirName)

	logger := p.baseLogger.With().Int("dump_counter", p.dumpCounter).Str("run_type", string(runType)).Logger()

	if err := os.MkdirAll(finalFolder, os.ModePerm); err != nil {
		return "", errors.WithStack(err)
	}

	logger.Info().Msgf("[Profiler] Starting dump process on folder %v", finalFolder)

	if err := p.cleanupFolder(baseDirectory, logger); err != nil {
		logger.Err(err).Send()
	}

	var finalErr error

	go func() {
		if err := p.collectStats(finalFolder); err != nil {
			logger.Err(err).Msg("[Profiler] Can not write ps command output")
			finalErr = err
		}

		if err := p.collectIoStat(finalFolder); err != nil {
			logger.Err(err).Msg("[Profiler] Can not write pidstat command output")
			finalErr = err
		}
	}()

	if err := p.makeCpuProfile(finalFolder, logger); err != nil {
		logger.Err(err).Msg("[Profiler] Can not write CPU dump")
		finalErr = err
	}

	if err := p.makeMemProfile(finalFolder, logger); err != nil {
		logger.Err(err).Msg("[Profiler] Can not write CPU dump")
		finalErr = err
	}

	p.lastProfiledAt = time.Now().UTC()

	return dirName, finalErr
}

func (p *Profiler) makeMemProfile(finalDir string, logger zerolog.Logger) error {
	heapProfileFile := path.Join(finalDir, "heap.pprof")
	allocsProfileFile := path.Join(finalDir, "allocs.pprof")

	fileHeap, err := os.Create(heapProfileFile)

	if err != nil {
		return errors.WithStack(err)
	}

	fileAllocs, err := os.Create(allocsProfileFile)

	if err != nil {
		return errors.WithStack(err)
	}

	defer func(fl *os.File, fl1 *os.File) {
		_ = fl.Close()
		_ = fl1.Close()
	}(fileHeap, fileAllocs)

	logger.Info().Msg("[Profiler] Starting runtime.GC()")

	runtime.GC() // get up-to-date statistics

	logger.Info().Msg("[Profiler] Writing Heap dump")

	if err := pprof.WriteHeapProfile(fileHeap); err != nil {
		return errors.WithStack(err)
	}

	writer := bufio.NewWriter(fileAllocs)

	if err := pprof.Lookup("allocs").WriteTo(writer, 0); err != nil {
		return errors.WithStack(err)
	}

	_ = writer.Flush()

	return nil
}

func (p *Profiler) makeCpuProfile(finalDir string, logger zerolog.Logger) error {
	cpuProfileFile := path.Join(finalDir, "cpu.pprof")

	fl, err := os.Create(cpuProfileFile)

	if err != nil {
		return errors.WithStack(err)
	}

	defer func(fl *os.File) {
		_ = fl.Close()
	}(fl)

	if err := pprof.StartCPUProfile(fl); err != nil {
		return errors.WithStack(err)
	}

	logger.Info().Msgf("[Profiler] Sleeping to collect CPU data for %v seconds", p.config.ProfileSessionLengthSeconds)
	time.Sleep(time.Duration(p.config.ProfileSessionLengthSeconds) * time.Second)

	pprof.StopCPUProfile()

	return nil
}

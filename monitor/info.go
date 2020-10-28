package monitor

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"os"
	"runtime"
)

type Info struct {
	Runtime RuntimeInfo
	Os      OsInfo
}

type RuntimeInfo struct {
	Pid     int
	Version string
}

type OsInfo struct {
	Name       string
	Platform   string
	Arch       string
	Version    string
	Processors int
}

func (s *Service) Info() Info {
	hostInfo, _ := host.Info()
	processors, _ := cpu.Counts(true)

	return Info{
		Runtime: RuntimeInfo{
			Pid:     os.Getpid(),
			Version: runtime.Version(),
		},
		Os: OsInfo{
			Name:       hostInfo.OS,
			Platform:   hostInfo.Platform + " " + hostInfo.PlatformVersion,
			Arch:       hostInfo.KernelArch,
			Version:    hostInfo.KernelVersion,
			Processors: processors,
		},
	}
}

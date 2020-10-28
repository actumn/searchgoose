package monitor

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"os"
	"runtime"
)

type Stats struct {
	Runtime RuntimeStats
	Os      OsStats
	Fs      FsStats
	Proc    ProcStats
}

type RuntimeStats struct {
	HeapAlloc uint64
	HeapSys   uint64
}

type OsStats struct {
	Cpu OsCpuStats
	Mem OsMemStats
}
type OsCpuStats struct {
	Percent     float64
	LoadAverage OsCpuLoad
}
type OsCpuLoad struct {
	Load1  float64
	Load5  float64
	Load15 float64
}
type OsMemStats struct {
	Total uint64
	Free  uint64
}

type FsStats struct {
	Total     uint64
	Free      uint64
	Available uint64
}

type ProcStats struct {
	CpuPercent      float64
	MemTotalVirtual uint64
	NumFDs          int32
}

func (s *Service) Stats() Stats {
	//cpuinfo, _ := cpu.Info()
	//fmt.Println(cpuinfo)

	percentages, _ := cpu.Percent(0, false)
	//fmt.Println(percentages[0])

	loadAvg, _ := load.Avg()

	//hostInfo, _ := host.Info()
	//fmt.Println(hostInfo)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memstats, _ := mem.VirtualMemory()
	//fmt.Println(memstats)

	diskStats, _ := disk.Usage("./data")
	//fmt.Println(diskStats)

	proc, _ := process.NewProcess(int32(os.Getpid()))
	procFd, _ := proc.NumFDs()
	procPercent, _ := proc.Percent(0)

	return Stats{
		Runtime: RuntimeStats{
			HeapAlloc: m.HeapAlloc,
			HeapSys:   m.HeapSys,
		},
		Os: OsStats{
			Cpu: OsCpuStats{
				Percent: percentages[0],
				LoadAverage: OsCpuLoad{
					Load1:  loadAvg.Load1,
					Load5:  loadAvg.Load5,
					Load15: loadAvg.Load15,
				},
			},
			Mem: OsMemStats{
				Total: memstats.Total,
				Free:  memstats.Free,
			},
		},
		Fs: FsStats{
			Total:     diskStats.Total,
			Free:      diskStats.Total - diskStats.Used,
			Available: diskStats.Free,
		},
		Proc: ProcStats{
			NumFDs:          procFd,
			CpuPercent:      procPercent,
			MemTotalVirtual: memstats.VMallocTotal,
		},
	}
}

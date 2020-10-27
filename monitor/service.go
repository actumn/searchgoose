package monitor

import (
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"runtime"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

type Stats struct {
	MemStats MemStats
	OsStats  OsStats
	FsStats  FsStats
}
type MemStats struct {
	Alloc uint64
	Sys   uint64
}
type OsStats struct {
	MemTotal uint64
	MemFree  uint64
}
type FsStats struct {
	Total     uint64
	Free      uint64
	Available uint64
}

func (s *Service) Stats() Stats {
	//cpuinfo, _ := cpu.Info()
	//fmt.Println(cpuinfo)
	//
	//percentages, _ := cpu.Percent(0, false)
	//fmt.Println(percentages[0])
	//
	//hostInfo, _ := host.Info()
	//fmt.Println(hostInfo)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memstats, _ := mem.VirtualMemory()
	//fmt.Println(memstats)

	diskStats, _ := disk.Usage("./data")
	//fmt.Println(diskStats)

	return Stats{
		MemStats: MemStats{
			Alloc: m.Alloc,
			Sys:   m.Sys,
		},
		OsStats: OsStats{
			MemTotal: memstats.Total,
			MemFree:  memstats.Free,
		},
		FsStats: FsStats{
			Total:     diskStats.Total,
			Free:      diskStats.Total - diskStats.Used,
			Available: diskStats.Free,
		},
		//	map[string]interface{}{
		//	"total_in_bytes": memstats.Total,
		//	"free_in_bytes": memstats.Free,
		//	//"used_in_bytes": memstats.Total - memstats.Free,
		//	//"free_percent": memstats.Free * 100 / memstats.Total, // 1
		//	//"used_percent": 100 - memstats.Free * 100 / memstats.Total, // 99
		//},
	}
}

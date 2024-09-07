package core

import (
	"fmt"
	"math"
	"unicode/utf8"

	"github.com/shirou/gopsutil/v4/disk"
)

type DiskInfo struct {
	PartitionInfo disk.PartitionStat
	Counters      disk.IOCountersStat
	Usage         disk.UsageStat
}

func (di *DiskInfo) String() string {
	return fmt.Sprintf("%s\n%s\n%s\n", di.PartitionInfo, di.Counters, di.Usage)
}

func DisksInfo() ([]*DiskInfo, error) {
	res := make([]*DiskInfo, 0)
	partitionsStat, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}
	wasNames := make(map[string]bool, 0)
	for _, ps := range partitionsStat {
		m, err := disk.IOCounters(ps.Device)
		if err != nil {
			return nil, err
		}
		for name, stat := range m {
			if !wasNames[name] {
				wasNames[name] = true
				if stat.Label != "" {
					u, err := disk.Usage(ps.Mountpoint)
					if err != nil {
						return nil, err
					}
					di := &DiskInfo{PartitionInfo: ps, Counters: stat, Usage: *u}
					res = append(res, di)
				}
			}
		}
	}
	return res, nil
}

func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func ListDisks(disksInfo []*DiskInfo) {
	rows := [][]string{
		{"Label", "Fs", "Size", "Used", "Available", "%Used", "Mountpoint"},
	}
	maxes := make([]int, len(rows[0]))
	for _, di := range disksInfo {
		rows = append(rows, []string{di.Counters.Label, di.PartitionInfo.Fstype, ByteCountIEC(int64(di.Usage.Total)), ByteCountIEC(int64(di.Usage.Used)), ByteCountIEC(int64(di.Usage.Free)), fmt.Sprintf("%.1f%%", di.Usage.UsedPercent), di.PartitionInfo.Mountpoint})
	}
	for _, row := range rows {
		for colNo, str := range row {
			l := utf8.RuneCountInString(str)
			maxes[colNo] = int(math.Max(float64(maxes[colNo]), float64(l)))
		}
	}
	for rowNo, row := range rows {
		for colNo, l := range maxes {
			format := fmt.Sprintf("%%-%ds ", l)
			if rowNo > 0 && (colNo > 0) && (colNo < len(maxes)-1) {
				format = fmt.Sprintf("%%%ds ", l)
			}
			fmt.Printf(format, row[colNo])
		}
		fmt.Println()
	}
}

package pcommon

import (
	"fmt"
	"runtime"
)

// see: https://golang.org/pkg/runtime/#MemStats
func PrintMemStats(head string) uint64 {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	fmt.Printf("PrintMemStats Memory %s\n", head)
	fmt.Printf("Memory Alloc = %v MiB\n", bToMb(stats.Alloc))
	fmt.Printf("Memory TotalAlloc = %v MiB\n", bToMb(stats.TotalAlloc))
	fmt.Printf("Memory HeapInuse = %v MiB\n", bToMb(stats.HeapInuse))
	fmt.Printf("Memory HeapIdle = %v MiB\n", bToMb(stats.HeapIdle))
	fmt.Printf("Memory StackInuse = %v MiB\n", bToMb(stats.StackInuse))
	fmt.Printf("Memory StackSys = %v MiB\n", bToMb(stats.StackSys))
	fmt.Printf("Memory Sys = %v MiB\n", bToMb(stats.Sys))
	fmt.Printf("Memory NumGC = %v\n", stats.NumGC)
	return bToMb(stats.Alloc)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

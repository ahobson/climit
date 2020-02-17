package main


import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
)

func readIntFromFile(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		return strconv.Atoi(scanner.Text())
	}
	return 0, fmt.Errorf("Empty file: %s", path)
}

const (
	CPU_QUOTA_FILE = "/sys/fs/cgroup/cpu/cpu.cfs_quota_us"
	CPU_PERIOD_FILE = "/sys/fs/cgroup/cpu/cpu.cfs_period_us"
	CPU_SHARES_FILE = "/sys/fs/cgroup/cpu/cpu.shares"

	PER_CPU_SHARES = 1024
)

// implementation ideas taken from the openjdk implementation
// https://github.com/openjdk/jdk/blob/d5686b87f31d6c57ec6b3e5e9c85a04209dbac7a/src/hotspot/os/linux/osContainer_linux.cpp#L487
func climitNproc(preferQuota bool) (int, error) {
	var quotaCount int = 0
	var shareCount int = 0
	var cpuCount = runtime.NumCPU()
	var limitCount int = 0

	quota, err := readIntFromFile(CPU_QUOTA_FILE)
	if err != nil {
		quota = -1
	}
	period, err := readIntFromFile(CPU_PERIOD_FILE)
	if err != nil {
		period = -1
	}
	share, err := readIntFromFile(CPU_SHARES_FILE)
	if err != nil {
		share = -1
	}

	if (quota > -1 && period > 0) {
		quotaCount = int(math.Ceil(float64(quota) / float64(period)))
	}

	if (share > -1) {
		shareCount = int(math.Ceil(float64(share) / float64(PER_CPU_SHARES)))
	}


	if (quotaCount != 0 && shareCount != 0) {
		if (preferQuota) {
			limitCount = quotaCount
		} else {
			if (shareCount < quotaCount) {
				limitCount = shareCount
			} else {
				limitCount = quotaCount
			}
		}
	} else if (quotaCount != 0) {
		limitCount = quotaCount
	} else {
		limitCount = shareCount
	}

	if (limitCount < cpuCount) {
		return limitCount, nil
	}

	return cpuCount, nil
}

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s nproc\n", os.Args[0])
		os.Exit(2)
	}
	nprocCommand := flag.NewFlagSet("nproc", flag.ExitOnError)
	preferQuota := nprocCommand.Bool("preferQuota", true, "prefer quota to shares")
	switch os.Args[1] {
	case "nproc":
		nprocCommand.Parse(os.Args[2:])
		nproc, err := climitNproc(*preferQuota)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting nproc: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("%d\n", nproc)
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", os.Args[1])
		os.Exit(1)
	}
}

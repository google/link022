package monitoring

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/golang/glog"
	"github.com/google/gnxi/utils/xpath"
	"github.com/google/link022/agent/gnmi"
	"github.com/google/link022/agent/syscmd"
)

const (
	statesUpdateDelay  = 15 * time.Second
	systemClockTick    = 100
	physicalMemoryPath = "/access-points/access-point[hostname=link022-pi-ap]/system/memory/state/physical"
	cpuUsagePath       = "/access-points/access-point[hostname=link022-pi-ap]/system/cpus/cpu[index=$index]/state/total/instant"
	channelPath        = "/access-points/access-point[hostname=link022-pi-ap]/radios/radio[id=$id]/channel"
	widthPath          = "/access-points/access-point[hostname=link022-pi-ap]/radios/radio[id=$id]/channel-width"
	frequencyPath      = "/access-points/access-point[hostname=link022-pi-ap]/radios/radio[id=$id]/operating-frequency"
	txpowerPath        = "/access-points/access-point[hostname=link022-pi-ap]/radios/radio[id=$id]/transmit-power"
	selfMemPath        = "/access-points/access-point[hostname=link022-pi-ap]/system/processes/process[pid=$pid]/state/memory-usage"
	selfCPUPath        = "/access-points/access-point[hostname=link022-pi-ap]/system/processes/process[pid=$pid]/state/cpu-utilization"
)

var cmdRunner = syscmd.Runner()

// UpdateDeviceStatus peroidically collect AP device stats
// and update their corresponding nodes in OpenConfig Model tree.
func UpdateDeviceStatus(bkgdContext context.Context, gnmiServer *gnmi.Server) {
	for {
		select {
		case <-bkgdContext.Done():
			return
		case <-time.After(statesUpdateDelay):
		}

		if err := updateMemoryInfo(gnmiServer); err != nil {
			log.Errorf("Error in updating memory info: %v", err)
		}
		if err := updateCPUInfo(gnmiServer); err != nil {
			log.Errorf("Error in updating memory info: %v", err)
		}
		if err := updateAPInfo(gnmiServer); err != nil {
			log.Errorf("Error in updating AP info: %v", err)
		}
	}
}

func updateMemoryInfo(s *gnmi.Server) error {
	b, err := ioutil.ReadFile("/proc/meminfo")
	if err != nil {
		return err
	}
	memStr := string(b)
	reFree := regexp.MustCompile("MemTotal:\\s+(\\d+)")
	match := reFree.FindStringSubmatch(memStr)
	if len(match) != 2 {
		return errors.New("No Memory Free info in /proc/meminfo")
	}
	pbPath, err := xpath.ToGNMIPath(physicalMemoryPath)
	if err != nil {
		return err
	}
	physicalMemory, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return err
	}
	stateOpt := gnmi.GNXIStateOptGenerator(pbPath, uint64(physicalMemory*1024), gnmi.InternalUpdateState)
	if err = s.InternalUpdate(stateOpt); err != nil {
		return err
	}

	pid := os.Getpid()
	spid := fmt.Sprint(pid)
	filePath := fmt.Sprintf("/proc/%v/status", pid)
	b, err = ioutil.ReadFile(filePath)
	if err != nil {
		log.Errorf("failed open %v: %v", filePath, err)
		return err
	}
	memStr = string(b)
	reSelfMem := regexp.MustCompile("VmRSS:\\s+(\\d+)")
	match = reSelfMem.FindStringSubmatch(memStr)
	if len(match) != 2 {
		return fmt.Errorf("No Memory info in: %v", filePath)
	}
	p := strings.Replace(selfMemPath, "$pid", spid, 1)
	pbPath, err = xpath.ToGNMIPath(p)
	if err != nil {
		log.Errorf("convert %v to GNMI path failed: %v", p, err)
		return err
	}
	selfMemory, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		log.Errorf("failed convert string to int: %v", err)
		return err
	}
	stateOpt = gnmi.GNXIStateOptGenerator(pbPath, uint64(selfMemory*1024), gnmi.InternalUpdateState)
	if err = s.InternalUpdate(stateOpt); err != nil {
		log.Errorf("update state failed: %v", err)
		return err
	}

	return nil
}

func updateCPUInfo(s *gnmi.Server) error {
	pid := os.Getpid()
	spid := fmt.Sprint(pid)
	filePath := fmt.Sprintf("/proc/%v/stat", pid)
	b0, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Errorf("failed open %v: %v", filePath, err)
		return err
	}
	time.Sleep(1 * time.Second)
	b1, err := ioutil.ReadFile("/proc/%v/stat")
	if err != nil {
		log.Errorf("failed open %v: %v", filePath, err)
		return err
	}
	cpuStr0 := strings.Split(string(b0), " ")
	cpuStr1 := strings.Split(string(b1), " ")
	if len(cpuStr0) < 14 || len(cpuStr1) < 14 {
		return errors.New("cpu info not correct")
	}
	up0, err := strconv.ParseInt(cpuStr0[13], 10, 64)
	if err != nil {
		log.Errorf("failed convert string to int: %v", err)
		return err
	}
	up1, err := strconv.ParseInt(cpuStr1[13], 10, 64)
	if err != nil {
		log.Errorf("failed convert string to int: %v", err)
		return err
	}
	cpuinfo, err := ioutil.ReadFile("/proc/cpuinfo")
	cpuCount := strings.Count(string(cpuinfo), "processor")
	cpuUtil := (up1 - up0) / (systemClockTick * int64(cpuCount))
	p := strings.Replace(selfCPUPath, "$pid", spid, 1)
	pbPath, err := xpath.ToGNMIPath(p)
	if err != nil {
		log.Errorf("convert %v to GNMI path failed: %v", p, err)
		return err
	}
	stateOpt := gnmi.GNXIStateOptGenerator(pbPath, uint8(cpuUtil), gnmi.InternalUpdateState)
	if err = s.InternalUpdate(stateOpt); err != nil {
		log.Errorf("update state failed: %v", err)
		return err
	}

	return nil
}

func updateAPInfo(s *gnmi.Server) error {
	//(TODO:tianyangz):add implementation
	return nil
}

package webapi

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	psnet "github.com/shirou/gopsutil/v4/net"
)

type SensorInfo struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
	Type  string  `json:"type"`
}

func getSensors() ([]SensorInfo, error) {
	var sensors []SensorInfo

	addSysInfo(&sensors)
	addCPUSensors(&sensors)
	addMemorySensors(&sensors)
	addDiskSensors(&sensors)
	addNetworkSensors(&sensors)
	addHwmonSensors(&sensors)
	addThermalSensors(&sensors)
	addVMwareSensors(&sensors)
	addLmSensors(&sensors)
	addGPUSensors(&sensors)
	addLoadSensors(&sensors)

	return sensors, nil
}

func addSysInfo(sensors *[]SensorInfo) {
	hostInfo, err := host.Info()
	if err != nil {
		return
	}
	*sensors = append(*sensors, SensorInfo{Name: "Hostname", Value: 0, Unit: hostInfo.Hostname, Type: "info"})
	*sensors = append(*sensors, SensorInfo{Name: "OS", Value: 0, Unit: hostInfo.Platform + " " + hostInfo.PlatformVersion, Type: "info"})
	*sensors = append(*sensors, SensorInfo{Name: "Kernel", Value: 0, Unit: hostInfo.KernelVersion, Type: "info"})
	*sensors = append(*sensors, SensorInfo{Name: "Uptime", Value: float64(hostInfo.Uptime), Unit: "seconds", Type: "uptime"})
}

func addCPUSensors(sensors *[]SensorInfo) {
	cpuInfos, err := cpu.Info()
	if err == nil && len(cpuInfos) > 0 {
		*sensors = append(*sensors, SensorInfo{Name: "CPU Model", Value: 0, Unit: cpuInfos[0].ModelName, Type: "info"})
		*sensors = append(*sensors, SensorInfo{Name: "CPU Cores (Logical)", Value: float64(len(cpuInfos)), Unit: "cores", Type: "info"})
		*sensors = append(*sensors, SensorInfo{Name: "CPU MHz", Value: cpuInfos[0].Mhz, Unit: "MHz", Type: "frequency"})

		physicalCores, _ := cpu.Counts(false)
		if physicalCores > 0 {
			*sensors = append(*sensors, SensorInfo{Name: "CPU Cores (Physical)", Value: float64(physicalCores), Unit: "cores", Type: "info"})
		}
	}

	cpuPercent, err := cpu.Percent(0, false)
	if err == nil && len(cpuPercent) > 0 {
		*sensors = append(*sensors, SensorInfo{Name: "CPU Usage", Value: cpuPercent[0], Unit: "%", Type: "usage"})
	}

	perCoreUsage, err := cpu.Percent(0, true)
	if err == nil {
		for i, pct := range perCoreUsage {
			if i >= 16 {
				break
			}
			*sensors = append(*sensors, SensorInfo{
				Name: fmt.Sprintf("CPU Core %d Usage", i), Value: pct, Unit: "%", Type: "usage",
			})
		}
	}

	readCPUFreq(sensors)
	readCPUTemperature(sensors)
}

func readCPUFreq(sensors *[]SensorInfo) {
	files, err := os.ReadDir("/sys/devices/system/cpu/cpufreq")
	if err == nil {
		for _, f := range files {
			if !f.IsDir() {
				continue
			}
			policyPath := "/sys/devices/system/cpu/cpufreq/" + f.Name()
			freqData, err := os.ReadFile(policyPath + "/scaling_cur_freq")
			if err != nil {
				continue
			}
			val, err := strconv.ParseFloat(strings.TrimSpace(string(freqData)), 64)
			if err != nil {
				continue
			}
			*sensors = append(*sensors, SensorInfo{
				Name: fmt.Sprintf("CPU Freq (%s)", f.Name()), Value: val / 1000.0, Unit: "MHz", Type: "frequency",
			})
		}
	}

	if len(*sensors) == 0 {
		for i := 0; i < 4; i++ {
			freqPath := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/scaling_cur_freq", i)
			data, err := os.ReadFile(freqPath)
			if err != nil {
				break
			}
			val, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
			if err != nil {
				continue
			}
			*sensors = append(*sensors, SensorInfo{
				Name: fmt.Sprintf("CPU%d Freq", i), Value: val / 1000.0, Unit: "MHz", Type: "frequency",
			})
		}
	}
}

func readCPUTemperature(sensors *[]SensorInfo) {
	for i := 0; i < 8; i++ {
		data, err := os.ReadFile(fmt.Sprintf("/sys/class/thermal/thermal_zone%d/temp", i))
		if err != nil {
			continue
		}
		val, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
		if err != nil {
			continue
		}
		typeData, _ := os.ReadFile(fmt.Sprintf("/sys/class/thermal/thermal_zone%d/type", i))
		zoneType := strings.TrimSpace(string(typeData))
		if zoneType == "" {
			zoneType = fmt.Sprintf("zone%d", i)
		}
		*sensors = append(*sensors, SensorInfo{
			Name: fmt.Sprintf("CPU Temp (%s)", zoneType), Value: val / 1000.0, Unit: "C", Type: "temperature",
		})
	}
}

func addMemorySensors(sensors *[]SensorInfo) {
	vmem, err := mem.VirtualMemory()
	if err == nil {
		*sensors = append(*sensors, SensorInfo{Name: "Memory Total", Value: float64(vmem.Total), Unit: "bytes", Type: "memory"})
		*sensors = append(*sensors, SensorInfo{Name: "Memory Used", Value: float64(vmem.Used), Unit: "bytes", Type: "memory"})
		*sensors = append(*sensors, SensorInfo{Name: "Memory Usage", Value: vmem.UsedPercent, Unit: "%", Type: "memory"})
		*sensors = append(*sensors, SensorInfo{Name: "Memory Available", Value: float64(vmem.Available), Unit: "bytes", Type: "memory"})
	}

	swap, err := mem.SwapMemory()
	if err == nil {
		*sensors = append(*sensors, SensorInfo{Name: "Swap Total", Value: float64(swap.Total), Unit: "bytes", Type: "memory"})
		*sensors = append(*sensors, SensorInfo{Name: "Swap Used", Value: float64(swap.Used), Unit: "bytes", Type: "memory"})
		*sensors = append(*sensors, SensorInfo{Name: "Swap Usage", Value: swap.UsedPercent, Unit: "%", Type: "memory"})
	}
}

func addDiskSensors(sensors *[]SensorInfo) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return
	}
	for _, p := range partitions {
		usage, err := disk.Usage(p.Mountpoint)
		if err != nil {
			continue
		}
		*sensors = append(*sensors, SensorInfo{
			Name: fmt.Sprintf("Disk %s Total", p.Mountpoint), Value: float64(usage.Total), Unit: "bytes", Type: "disk",
		})
		*sensors = append(*sensors, SensorInfo{
			Name: fmt.Sprintf("Disk %s Used", p.Mountpoint), Value: float64(usage.Used), Unit: "bytes", Type: "disk",
		})
		*sensors = append(*sensors, SensorInfo{
			Name: fmt.Sprintf("Disk %s Usage", p.Mountpoint), Value: usage.UsedPercent, Unit: "%", Type: "disk",
		})
	}

	entries, err := os.ReadDir("/sys/block")
	if err != nil {
		return
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "loop") || strings.HasPrefix(entry.Name(), "ram") {
			continue
		}
		statPath := "/sys/block/" + entry.Name() + "/stat"
		data, err := os.ReadFile(statPath)
		if err != nil {
			continue
		}
		fields := strings.Fields(string(data))
		if len(fields) >= 10 {
			readIO, _ := strconv.ParseFloat(fields[0], 64)
			writeIO, _ := strconv.ParseFloat(fields[4], 64)
			*sensors = append(*sensors, SensorInfo{
				Name: fmt.Sprintf("Disk %s Read Ops", entry.Name()), Value: readIO, Unit: "ops", Type: "disk_io",
			})
			*sensors = append(*sensors, SensorInfo{
				Name: fmt.Sprintf("Disk %s Write Ops", entry.Name()), Value: writeIO, Unit: "ops", Type: "disk_io",
			})
		}
	}
}

func addNetworkSensors(sensors *[]SensorInfo) {
	netIO, err := psnet.IOCounters(true)
	if err != nil {
		return
	}
	for _, nic := range netIO {
		if nic.Name == "lo" {
			continue
		}
		*sensors = append(*sensors, SensorInfo{
			Name: fmt.Sprintf("NIC %s RX Bytes", nic.Name), Value: float64(nic.BytesRecv), Unit: "bytes", Type: "network",
		})
		*sensors = append(*sensors, SensorInfo{
			Name: fmt.Sprintf("NIC %s TX Bytes", nic.Name), Value: float64(nic.BytesSent), Unit: "bytes", Type: "network",
		})
		*sensors = append(*sensors, SensorInfo{
			Name: fmt.Sprintf("NIC %s RX Packets", nic.Name), Value: float64(nic.PacketsRecv), Unit: "packets", Type: "network",
		})
		*sensors = append(*sensors, SensorInfo{
			Name: fmt.Sprintf("NIC %s TX Packets", nic.Name), Value: float64(nic.PacketsSent), Unit: "packets", Type: "network",
		})
		*sensors = append(*sensors, SensorInfo{
			Name: fmt.Sprintf("NIC %s RX Errors", nic.Name), Value: float64(nic.Errin), Unit: "errors", Type: "network",
		})
		*sensors = append(*sensors, SensorInfo{
			Name: fmt.Sprintf("NIC %s TX Errors", nic.Name), Value: float64(nic.Errout), Unit: "errors", Type: "network",
		})
		if nic.Name == "ens33" || len(netIO) <= 3 {
		}
	}
}

func addHwmonSensors(sensors *[]SensorInfo) {
	entries, err := os.ReadDir("/sys/class/hwmon")
	if err != nil {
		return
	}

	for _, entry := range entries {
		hwmonPath := "/sys/class/hwmon/" + entry.Name()
		nameBytes, _ := os.ReadFile(hwmonPath + "/name")
		chipName := strings.TrimSpace(string(nameBytes))
		if chipName == "" {
			chipName = entry.Name()
		}

		files, err := os.ReadDir(hwmonPath)
		if err != nil {
			continue
		}

		hasSensors := false
		for _, file := range files {
			fname := file.Name()

			if strings.HasPrefix(fname, "temp") && strings.HasSuffix(fname, "_input") {
				label := readHwmonLabel(hwmonPath, fname)
				data, err := os.ReadFile(hwmonPath + "/" + fname)
				if err != nil {
					continue
				}
				val, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
				if err != nil {
					continue
				}
				*sensors = append(*sensors, SensorInfo{
					Name: chipName + " " + label, Value: val / 1000.0, Unit: "C", Type: "temperature",
				})
				hasSensors = true
			}

			if strings.HasPrefix(fname, "fan") && strings.HasSuffix(fname, "_input") {
				label := readHwmonLabel(hwmonPath, fname)
				data, err := os.ReadFile(hwmonPath + "/" + fname)
				if err != nil {
					continue
				}
				val, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
				if err != nil {
					continue
				}
				*sensors = append(*sensors, SensorInfo{
					Name: chipName + " " + label, Value: val, Unit: "RPM", Type: "fan",
				})
				hasSensors = true
			}

			if (strings.HasPrefix(fname, "in") || strings.HasPrefix(fname, "curr")) && strings.HasSuffix(fname, "_input") {
				label := readHwmonLabel(hwmonPath, fname)
				data, err := os.ReadFile(hwmonPath + "/" + fname)
				if err != nil {
					continue
				}
				val, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
				if err != nil {
					continue
				}
				*sensors = append(*sensors, SensorInfo{
					Name: chipName + " " + label, Value: val / 1000.0, Unit: "V", Type: "voltage",
				})
				hasSensors = true
			}
		}

		if !hasSensors && chipName != "ACAD" {
			*sensors = append(*sensors, SensorInfo{
				Name: chipName, Value: 0, Unit: "(no sensor data)", Type: "info",
			})
		}
	}
}

func readHwmonLabel(hwmonPath, inputName string) string {
	baseName := strings.TrimSuffix(inputName, "_input")
	labelPath := hwmonPath + "/" + baseName + "_label"
	data, err := os.ReadFile(labelPath)
	if err == nil {
		label := strings.TrimSpace(string(data))
		if label != "" {
			return label
		}
	}
	return baseName
}

func addThermalSensors(sensors *[]SensorInfo) {
	entries, err := os.ReadDir("/sys/class/thermal")
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "thermal_zone") {
			continue
		}

		thermalPath := "/sys/class/thermal/" + entry.Name()
		typeBytes, _ := os.ReadFile(thermalPath + "/type")
		zoneType := strings.TrimSpace(string(typeBytes))
		if zoneType == "" {
			zoneType = entry.Name()
		}

		tempBytes, err := os.ReadFile(thermalPath + "/temp")
		if err != nil {
			continue
		}
		val, err := strconv.ParseFloat(strings.TrimSpace(string(tempBytes)), 64)
		if err != nil {
			continue
		}

		*sensors = append(*sensors, SensorInfo{
			Name: "Thermal " + zoneType, Value: val / 1000.0, Unit: "C", Type: "temperature",
		})
	}
}

func addVMwareSensors(sensors *[]SensorInfo) {
	cmd := exec.Command("vmware-toolbox-cmd", "stat", "speed")
	out, err := cmd.Output()
	if err == nil {
		val, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
		if err == nil && val > 0 {
			*sensors = append(*sensors, SensorInfo{
				Name: "CPU Frequency (Host)", Value: val, Unit: "MHz", Type: "frequency",
			})
		}
	}

	cmd = exec.Command("vmware-toolbox-cmd", "stat", "memlimit")
	out, err = cmd.Output()
	if err == nil {
		val, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
		if err == nil && val > 0 {
			*sensors = append(*sensors, SensorInfo{
				Name: "VM Memory Limit", Value: val, Unit: "MB", Type: "memory",
			})
		}
	}

	cmd = exec.Command("vmware-toolbox-cmd", "stat", "memres")
	out, err = cmd.Output()
	if err == nil {
		val, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
		if err == nil && val > 0 {
			*sensors = append(*sensors, SensorInfo{
				Name: "VM Memory Reserved", Value: val, Unit: "MB", Type: "memory",
			})
		}
	}

	cmd = exec.Command("vmware-toolbox-cmd", "stat", "balloon")
	out, err = cmd.Output()
	if err == nil {
		val, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
		if err == nil && val > 0 {
			*sensors = append(*sensors, SensorInfo{
				Name: "VM Ballooned Memory", Value: val, Unit: "MB", Type: "memory",
			})
		}
	}

	cmd = exec.Command("vmware-toolbox-cmd", "stat", "sessionid")
	out, err = cmd.Output()
	if err == nil {
		sessionID := strings.TrimSpace(string(out))
		if sessionID != "" {
			*sensors = append(*sensors, SensorInfo{
				Name: "VM Session ID", Value: 0, Unit: sessionID, Type: "info",
			})
		}
	}
}

func addLmSensors(sensors *[]SensorInfo) {
	cmd := exec.Command("sensors", "-u")
	out, err := cmd.Output()
	if err != nil {
		return
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	var currentChip, currentLabel string

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "  ") {
			currentChip = strings.TrimSpace(line)
			continue
		}

		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "temp") && strings.HasSuffix(trimmed, "_input:") {
			parts := strings.Split(trimmed, ":")
			if len(parts) >= 2 {
				currentLabel = strings.TrimSpace(parts[0])
				currentLabel = strings.TrimSuffix(currentLabel, "_input")
			}
			continue
		}

		if strings.HasPrefix(trimmed, "fan") && strings.HasSuffix(trimmed, "_input:") {
			parts := strings.Split(trimmed, ":")
			if len(parts) >= 2 {
				currentLabel = strings.TrimSpace(parts[0])
				currentLabel = strings.TrimSuffix(currentLabel, "_input")
			}
			continue
		}

		if strings.HasPrefix(trimmed, "temp") || strings.HasPrefix(trimmed, "fan") ||
			strings.HasPrefix(trimmed, "in") || strings.HasPrefix(trimmed, "curr") {
			if strings.Contains(trimmed, ":") && strings.Contains(trimmed, "_input") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					valStr := strings.TrimSpace(parts[1])
					val, err := strconv.ParseFloat(valStr, 64)
					if err == nil {
						label := currentLabel
						if label == "" {
							label = strings.TrimSpace(parts[0])
						}
						unit := ""
						sensorType := "unknown"
						if strings.HasPrefix(strings.TrimSpace(line), "temp") {
							sensorType = "temperature"
							unit = "C"
						} else if strings.HasPrefix(strings.TrimSpace(line), "fan") {
							sensorType = "fan"
							unit = "RPM"
						} else if strings.HasPrefix(strings.TrimSpace(line), "in") || strings.HasPrefix(strings.TrimSpace(line), "curr") {
							sensorType = "voltage"
							unit = "V"
						}
						*sensors = append(*sensors, SensorInfo{
							Name: currentChip + " " + label, Value: val, Unit: unit, Type: sensorType,
						})
					}
				}
			}
		}
	}
}

func addGPUSensors(sensors *[]SensorInfo) {
	entries, err := os.ReadDir("/sys/class/drm")
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !strings.HasPrefix(entry.Name(), "card") {
			continue
		}

		cardPath := "/sys/class/drm/" + entry.Name() + "/device"

		tempPath := cardPath + "/hwmon/hwmon*/temp1_input"
		matches, _ := os.ReadDir("/sys/class/drm/" + entry.Name() + "/device/hwmon/")
		for _, m := range matches {
			if strings.HasPrefix(m.Name(), "hwmon") {
				tempFile := "/sys/class/drm/" + entry.Name() + "/device/hwmon/" + m.Name() + "/temp1_input"
				data, err := os.ReadFile(tempFile)
				if err == nil {
					val, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
					if err == nil {
						*sensors = append(*sensors, SensorInfo{
							Name: fmt.Sprintf("GPU %s Temp", entry.Name()), Value: val / 1000.0, Unit: "C", Type: "temperature",
						})
					}
				}
			}
		}
		_ = tempPath
	}

	nvidiaCmd := exec.Command("nvidia-smi", "--query-gpu=temperature.gpu,fan.speed,utilization.gpu,memory.used,memory.total", "--format=csv,noheader,nounits")
	out, err := nvidiaCmd.Output()
	if err == nil {
		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		gpuIdx := 0
		for scanner.Scan() {
			fields := strings.Split(scanner.Text(), ", ")
			if len(fields) >= 3 {
				temp, _ := strconv.ParseFloat(fields[0], 64)
				fanSpeed, _ := strconv.ParseFloat(fields[1], 64)
				gpuUtil, _ := strconv.ParseFloat(fields[2], 64)
				*sensors = append(*sensors, SensorInfo{
					Name: fmt.Sprintf("GPU%d Temperature", gpuIdx), Value: temp, Unit: "C", Type: "temperature",
				})
				*sensors = append(*sensors, SensorInfo{
					Name: fmt.Sprintf("GPU%d Fan Speed", gpuIdx), Value: fanSpeed, Unit: "%", Type: "fan",
				})
				*sensors = append(*sensors, SensorInfo{
					Name: fmt.Sprintf("GPU%d Usage", gpuIdx), Value: gpuUtil, Unit: "%", Type: "usage",
				})
				if len(fields) >= 5 {
					memUsed, _ := strconv.ParseFloat(fields[3], 64)
					memTotal, _ := strconv.ParseFloat(fields[4], 64)
					*sensors = append(*sensors, SensorInfo{
						Name: fmt.Sprintf("GPU%d Mem Used", gpuIdx), Value: memUsed, Unit: "MiB", Type: "memory",
					})
					*sensors = append(*sensors, SensorInfo{
						Name: fmt.Sprintf("GPU%d Mem Total", gpuIdx), Value: memTotal, Unit: "MiB", Type: "memory",
					})
				}
			}
			gpuIdx++
		}
	}
}

func addLoadSensors(sensors *[]SensorInfo) {
	loadAvg, err := load.Avg()
	if err != nil {
		return
	}
	*sensors = append(*sensors, SensorInfo{Name: "Load 1min", Value: loadAvg.Load1, Unit: "", Type: "load"})
	*sensors = append(*sensors, SensorInfo{Name: "Load 5min", Value: loadAvg.Load5, Unit: "", Type: "load"})
	*sensors = append(*sensors, SensorInfo{Name: "Load 15min", Value: loadAvg.Load15, Unit: "", Type: "load"})
}

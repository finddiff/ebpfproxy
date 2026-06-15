package webapi

import (
	"bufio"
	"os"
	"strings"
)

type DHCPLease struct {
	IP        string `json:"ip"`
	MAC       string `json:"mac"`
	Hostname  string `json:"hostname"`
	ExpiresAt string `json:"expires_at"`
}

func getDHCPLeases() ([]DHCPLease, error) {
	var leases []DHCPLease

	leaseFiles := []string{
		"/var/lib/dhcp/dhcpd.leases",
		"/var/lib/dhcpd/dhcpd.leases",
		"/var/lib/misc/dnsmasq.leases",
		"/tmp/dhcp.leases",
	}

	for _, path := range leaseFiles {
		if l, err := parseDHCPLeases(path); err == nil && len(l) > 0 {
			leases = append(leases, l...)
		}
	}

	if len(leases) > 0 {
		return leases, nil
	}

	leases = readDnsmasqLeases()
	return leases, nil
}

func parseDHCPLeases(path string) ([]DHCPLease, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var leases []DHCPLease
	scanner := bufio.NewScanner(file)

	var currentLease DHCPLease
	inLease := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "lease ") {
			inLease = true
			currentLease = DHCPLease{}
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				currentLease.IP = parts[1]
			}
		} else if inLease && strings.HasPrefix(line, "}") {
			inLease = false
			if currentLease.IP != "" && currentLease.MAC != "" {
				leases = append(leases, currentLease)
			}
		} else if inLease {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				switch strings.TrimSuffix(fields[0], ";") {
				case "hardware":
					if len(fields) >= 3 {
						currentLease.MAC = strings.TrimSuffix(fields[2], ";")
					}
				case "client-hostname":
					currentLease.Hostname = strings.Trim(strings.TrimSuffix(fields[1], ";"), "\"")
				case "ends":
					currentLease.ExpiresAt = strings.Join(fields[1:], " ")
					currentLease.ExpiresAt = strings.TrimSuffix(currentLease.ExpiresAt, ";")
				}
			}
		}
	}

	return leases, nil
}

func readDnsmasqLeases() []DHCPLease {
	var leases []DHCPLease

	paths := []string{
		"/var/lib/misc/dnsmasq.leases",
		"/tmp/dhcp.leases",
	}

	for _, path := range paths {
		file, err := os.Open(path)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			parts := strings.Fields(line)
			if len(parts) >= 4 {
				leases = append(leases, DHCPLease{
					ExpiresAt: parts[0],
					MAC:       parts[1],
					IP:        parts[2],
					Hostname:  parts[3],
				})
			}
		}
		file.Close()

		if len(leases) > 0 {
			break
		}
	}

	return leases
}

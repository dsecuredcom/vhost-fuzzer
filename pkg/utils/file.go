package utils

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func isValidIP(ip string) bool {
	parsedIP := net.ParseIP(strings.TrimSpace(ip))
	return parsedIP != nil && parsedIP.To4() != nil
}

func expandCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR notation: %s", cidr)
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); incrementIP(ip) {
		ips = append(ips, ip.String())
	}

	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}

	return ips, nil
}

func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func expandIPRange(ipRange string) ([]string, error) {
	if strings.Contains(ipRange, "/") {
		return expandCIDR(ipRange)
	}

	if !strings.Contains(ipRange, "-") {
		if !isValidIP(ipRange) {
			return nil, fmt.Errorf("invalid IP address: %s", ipRange)
		}
		return []string{strings.TrimSpace(ipRange)}, nil
	}

	parts := strings.Split(ipRange, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid IP range format: %s", ipRange)
	}

	startIP := net.ParseIP(strings.TrimSpace(parts[0]))
	endIP := net.ParseIP(strings.TrimSpace(parts[1]))

	if startIP == nil || endIP == nil || startIP.To4() == nil || endIP.To4() == nil {
		return nil, fmt.Errorf("invalid IP address in range: %s", ipRange)
	}

	start := ipToUint32(startIP)
	end := ipToUint32(endIP)

	if end < start {
		return nil, fmt.Errorf("invalid range: end IP is less than start IP")
	}

	if end-start > 65536 {
		return nil, fmt.Errorf("IP range too large: more than 65536 addresses")
	}

	var ips []string
	for i := start; i <= end; i++ {
		ip := uint32ToIP(i)
		ips = append(ips, ip.String())
	}

	return ips, nil
}

func ipToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}

func uint32ToIP(n uint32) net.IP {
	ip := make(net.IP, 4)
	ip[0] = byte(n >> 24)
	ip[1] = byte(n >> 16)
	ip[2] = byte(n >> 8)
	ip[3] = byte(n)
	return ip
}

func LoadIpLines(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filename, err)
		os.Exit(1)
	}
	defer file.Close()

	uniqueLines := make(map[string]struct{})
	var invalidLines []string

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}

		ips, err := expandIPRange(line)
		if err != nil {
			invalidLines = append(invalidLines, fmt.Sprintf("Line %d: %s - %v", lineNum, line, err))
			continue
		}

		for _, ip := range ips {
			uniqueLines[ip] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		os.Exit(1)
	}

	if len(invalidLines) > 0 {
		fmt.Printf("\nWarning: Found invalid entries in %s:\n", filename)
		for _, line := range invalidLines {
			fmt.Printf("  %s\n", line)
		}
		fmt.Println()
	}

	result := make([]string, 0, len(uniqueLines))
	for line := range uniqueLines {
		result = append(result, line)
	}

	return result
}

func LoadHostLines(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filename, err)
		os.Exit(1)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		os.Exit(1)
	}

	return lines
}

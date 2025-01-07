# vhost-fuzzer

A high-performance virtual host fuzzing tool designed to discover virtual hosts by testing different host headers against IP addresses. It supports concurrent scanning, custom paths, and flexible filtering options.

## Features

- Fast concurrent scanning with customizable worker count
- Support for both HTTP and HTTPS protocols
- Custom path testing
- Response filtering by status code and body content
- Efficient memory management with connection pooling
- Progress bar with real-time scanning status
- Verbose mode for detailed request/response inspection

## Building

### Prerequisites

- Go 1.19 or higher

### Installation

```bash
# Clone the repository
git clone https://github.com/dsecuredcom/vhost-fuzzer.git

# Change to the project directory
cd vhost-fuzzer

# Build the binary
go build
```

## Usage

Basic usage requires two input files: one containing IP addresses and another containing hostnames to test:

```bash
./vhost-fuzzer -ips ips.txt -hosts hosts.txt
```

### Input File Format

Both the IPs and hosts files should contain one entry per line:

**ips.txt:**
```
192.168.1.1
192.168.1.2
```

**hosts.txt:**
```
example.com
test.example.com
```

## Flags

| Flag                  | Default | Description                                     |
|-----------------------|---|-------------------------------------------------|
| `-ips`                | | File containing IP addresses (required)         |
| `-hosts`              | | File containing hostnames (required)            |
| `-concurrency`        | 100 | Number of concurrent workers                    |
| `-paths`              | "/" | Comma-separated list of paths to check          |
| `-protocol`           | "http" | Protocol to use (http/https)                    |
| `-http-body-includes` | | String to search for in response body           |
| `-http-status-is`     | 0 | Expected HTTP status code                       |
| `-request-timeout`    | 4 | Timeout for individual requests in seconds      |
| `-max-idle-timeout`   | 6 | Maximum idle connection duration in seconds     |
| `-max-conn-timeout`   | 6 | Maximum connection duration in seconds          |
| `-read-timeout`       | 5 | Read timeout in seconds                         |
| `-write-timeout`      | 5 | Write timeout in seconds                        |
| `-headers`            | | Additional HTTP headers (Format: "X1:v1; X2:v2) |
| `-host-filter`        | | Use only hosts with those strings (CSV)         |
| `-ip-filter`          | | Use only ips with those strings (CSV)           |
| `-verbose`            | false | Show all requests and responses                 |

## Examples

### Basic Scan
```bash
./vhost-fuzzer -ips ips.txt -hosts hosts.txt
```

### Advanced Usage
```bash
# Scan with HTTPS and custom paths
./vhost-fuzzer -ips ips.txt -hosts hosts.txt -protocol https -paths /,/admin,/api

# Scan with specific status code matching
./vhost-fuzzer -ips ips.txt -hosts hosts.txt -http-status-is 200

# High-concurrency scan with body content matching
./vhost-fuzzer -ips ips.txt -hosts hosts.txt -concurrency 200 -http-body-includes "Welcome"

# Verbose mode with custom timeouts
./vhost-fuzzer -ips ips.txt -hosts hosts.txt -verbose -request-timeout 10 -read-timeout 8
```

## Output

The tool will display:
1. Total number of targets to be scanned
2. Progress bar showing scanning status
3. Any matches found based on specified criteria
4. Scan duration upon completion

In verbose mode (`-verbose`), it will also show detailed request and response information for each attempt.

## Notes

- The tool automatically adjusts GOMAXPROCS to match the concurrency level
- All paths are automatically prefixed with "/" if not provided
- HTTPS connections skip certificate verification
- The progress bar updates every 10,000 requests or every second, whichever comes first
- Memory usage is optimized through connection and request/response pooling
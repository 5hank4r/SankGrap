# SankGrab

SankGrab is a lightweight and efficient CLI tool for extracting subdomains from HTTP response headers and bodies. Designed for speed and simplicity, it's perfect for researchers and developers working on subdomain enumeration.

## Features

- Extract subdomains from:
  - HTTP response headers (`-m rh`)
  - HTTP response bodies (`-m rb`)
  - Both headers and bodies (`-m both`)
- Fast and concurrent processing.
- Supports custom domain filtering using regex.
- Progress bar to track the process.
- Verbose logging for debugging.

## Installation

### Clone the Repository
```bash
git clone https://github.com/yourusername/SankGrab.git
cd SankGrab
````

### Flag	Description	Default
```
-d	Target domain to filter subdomains (e.g., example.com).	Required
-f	Input file containing the list of URLs to process.	Required
-o	Output file to save extracted subdomains. If not specified, prints to stdout.	None
-w	Number of concurrent workers for processing URLs.	10
-m	Extraction mode: rb (body), rh (header), or both.	both
-v	Enable verbose logging.	false
```


#### Extract from Both Headers and Body (Default)
```bash
./sankgrab -d example.com -f urls.txt -o results.txt -m both
```
### Extract Only from Response Headers
```bash
./sankgrab -d example.com -f urls.txt -m rh
```
### Extract Only from Response Bodies
```bash
./sankgrab -d example.com -f urls.txt -m rb
```

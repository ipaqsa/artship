# Artship

**A CLI tool for extracting and examining artifacts from OCI/Docker images**

Artship (artifact ship) enables engineers to efficiently extract specific files, binaries, or directories from container images without running containers. 
Suitable for CI/CD pipelines, artifact distribution, and container image analysis.

### Key Features

#### 🚀 **Container Image Management**
- Extract files, binaries, and directories from any OCI/Docker image
- **Mirror images** between registries without Docker daemon
- Support for public and private container registries
- Compatible with any registry
- No container runtime required - direct image layer access

#### 🔍 **Advanced Image Analysis**
- **Compare images** - see what changed between versions (diff command)
- **List artifacts** with filtering by type (files, dirs, symlinks)
- **Detailed information** display (size, permissions, type)
- **Content preview** of files directly from images
- **Image metadata** inspection (layers, architecture, environment, labels)
- **Repository exploration** - list all available tags

#### 🔐 **Security**
- **Authentication support** for private registries (username/password, token, auth string)
- **Docker credential integration** - seamless keychain support

### Installation

#### From GitHub Releases (Recommended)

Download pre-built binaries for your platform from the [latest release page](https://github.com/ipaqsa/artship/releases/latest). 

Binaries are available for:

- **Linux**: amd64, arm64
- **macOS**: arm64 (Apple Silicon)

```bash
# Example for Linux amd64
wget https://github.com/ipaqsa/artship/releases/latest/download/artship-linux-amd64
chmod +x artship-linux-amd64
sudo mv artship-linux-amd64 /usr/local/bin/artship

# Example for macOS arm64
wget https://github.com/ipaqsa/artship/releases/latest/download/artship-darwin-arm64
chmod +x artship-darwin-arm64
sudo mv artship-darwin-arm64 /usr/local/bin/artship
```

### From Source

```bash
git clone https://github.com/ipaqsa/artship
cd artship
make build
sudo make install
```

### Using Go

```bash
go install github.com/ipaqsa/artship/cmd/artship@latest
```

### Using Docker

```bash
# Run directly with Docker
docker run --rm -v $(pwd):/workspace ghcr.io/ipaqsa/artship:latest cp \
  nginx:latest \
  --artifact nginx \
  --output /workspace/nginx
```

## Quick Start

### Basic Usage

```bash
# List available artifacts in an image
artship ls nginx:latest

# Extract a specific binary
artship cp nginx:latest --artifact nginx --output ./nginx

# View file content without extraction
artship cat nginx:latest /etc/nginx/nginx.conf

# Compare two image versions
artship diff nginx:1.24 nginx:1.25

# Mirror image between registries
artship mirror nginx:latest myregistry.com/nginx:latest -u admin -p secret
```

### Common Commands

```bash
# List available tags for a repository
artship tags nginx
Available tags:
1
1-alpine
1-alpine-otel
1-alpine-perl
1-alpine-slim
# ... (many more tags)
```

```bash
# Copy artifacts from a container image
artship cp nginx:latest --artifact usr/sbin/nginx --output ./nginx-binary
Downloading image: nginx:latest
Successfully copied 1 artifacts:
  usr/sbin/nginx -> ./nginx-binary (file, 1.5 MB)
Total size: 1.5 MB
```

```bash
# List artifacts with detailed information
artship ls alpine:latest --detailed --filter file | head -5
TYPE     SIZE       MODE     PATH
-------- ---------- -------- --------
file     789.8 KB   0755     bin/busybox
file     7 B        0644     etc/alpine-release
file     7 B        0644     etc/apk/arch
```

```bash
# Show content of a specific file
artship cat nginx:latest /etc/nginx/nginx.conf
user  nginx;
worker_processes  auto;
error_log  /var/log/nginx/error.log notice;
# ... (rest of config)
```

```bash
# Display image metadata (including labels)
artship meta nginx:latest
Image: nginx:latest
Digest: sha256:f15190cd...
Architecture: amd64
OS: linux
Size: 2292 bytes
Layers: 7
Created: 2025-08-13T16:34:01Z

Environment Variables:
  PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
  NGINX_VERSION=1.29.1

Labels:
  maintainer=NGINX Docker Maintainers <docker-maint@nginx.com>
```

### Use Cases

#### 🏗️ **CI/CD Pipelines**
- Extract build artifacts from multi-stage Docker builds
- Deploy specific binaries without running containers
- Retrieve configuration files and assets for deployment
- Validate build outputs and inspect container contents

#### 📦 **Artifact Distribution**
- Distribute compiled binaries across different environments
- Share configuration templates from containerized applications
- Extract vendor dependencies and third-party tools
- Create lightweight deployment packages

#### 🔍 **Container Image Analysis**
- Inspect and audit container image contents
- Extract security certificates and configuration files
- Analyze image layers and file structures
- Reverse engineer containerized applications

#### Examples

##### Explore an image before copying
```bash
# First, see what's available in the image
artship ls nginx:latest
./
bin
boot/
dev/
etc/
usr/
usr/bin/
usr/sbin/
# ... (many more files)
```

```bash
# List with detailed information (size, type, permissions)
artship ls nginx:latest --detailed | head -10
TYPE     SIZE       MODE     PATH
-------- ---------- -------- --------
dir      0 B        0755     ./
symlink  -          0777     bin
dir      0 B        0755     boot/
dir      0 B        0755     dev/
dir      0 B        0755     etc/
file     0 B        0600     etc/.pwd.lock
file     3.0 KB     0644     etc/adduser.conf
```

```bash
# Filter by file type
artship ls nginx:latest --filter file | grep nginx
usr/sbin/nginx
usr/sbin/nginx-debug
etc/default/nginx
etc/init.d/nginx
etc/logrotate.d/nginx
```

```bash
# Check image metadata and structure
artship meta nginx:latest
Image: nginx:latest
Architecture: amd64
OS: linux
Layers: 7
Created: 2025-08-13T16:34:01Z
```

```bash
# View a configuration file
artship cat nginx:latest /etc/nginx/nginx.conf
user  nginx;
worker_processes  auto;
error_log  /var/log/nginx/error.log notice;
pid        /run/nginx.pid;
```

##### Copy a single binary from image
```bash
artship cp nginx:latest --artifact usr/sbin/nginx --output /usr/local/bin/nginx
Downloading image: nginx:latest
Extracted file to: /usr/local/bin/nginx
Successfully copied 1 artifacts:
  usr/sbin/nginx -> /usr/local/bin/nginx (file, 1.5 MB)
Total size: 1.5 MB
```

```bash
# Copy to current directory (uses artifact name as filename)
artship cp nginx:latest -a usr/sbin/nginx -o .
Downloading image: nginx:latest
Extracted file to: ./nginx
Successfully copied 1 artifacts:
  usr/sbin/nginx -> ./nginx (file, 1.5 MB)
Total size: 1.5 MB
```

##### Copy directories and files
```bash
artship cp myapp:latest \
  --artifact /app/bin \
  --artifact /app/config \
  --output ./local/bin \
  --output ./local/config
```

##### Copy from private registry
```bash
# Using credentials
artship cp my-registry.com/myapp:v1.0 \
  --artifact myapp \
  --output ./bin/myapp \
  --username myuser \
  --password mypass
```

```bash
# Using Docker credentials (automatic)
artship cp my-registry.com/myapp:v1.0 \
  --artifact myapp \
  --output ./bin/myapp
```

##### Extract entire image
```bash
# Extract all files from nginx image to current directory
artship extract nginx:latest

# Extract all files to a specific directory
artship extract alpine:latest --output ./extracted-alpine

# Extract from a private registry
artship extract my-registry.com/myapp:v1.0 --output ./extracted-app
```

##### Check artifact existence
```bash
# Check if nginx binary exists
artship has nginx:latest nginx

# Check for configuration file
artship has nginx:latest /etc/nginx/nginx.conf
```

##### Get detailed artifact information
```bash
# Show detailed info about nginx binary
artship info nginx:latest nginx

# Show info about configuration file
artship info nginx:latest /etc/nginx/nginx.conf
```

### Docker Build & Extract Examples

#### Example 1: Extracting Configuration Files

```dockerfile
FROM alpine:latest

# Create various config files
RUN mkdir -p /app/config /app/data /app/bin
COPY <<EOF /app/config/app.yml
server:
  port: 8080
  host: localhost
database:
  driver: postgres
  host: db
EOF

COPY <<EOF /app/config/nginx.conf
server {
    listen 80;
    location / { proxy_pass https://backend:3000; }
}
EOF

RUN echo '#!/bin/sh\necho "Starting app..."' > /app/bin/start.sh
RUN chmod +x /app/bin/start.sh
```

```bash
# Build and extract configurations
docker build -t config-image .

# List all configuration files
artship ls config-image -f file -d

# Extract all config files
artship cp config-image \
  -a /app/config/app.yml \
  -a /app/config/nginx.conf \
  -a /app/bin/start.sh \
  -o ./config/ \
  -o ./config/ \
  -o ./scripts/

# View extracted content
cat config/app.yml
cat scripts/start.sh
```

### Example 2: Private Registry with Authentication

```bash
# Build and push to private registry
docker build -t private-registry.company.com/my-app:latest .
docker push private-registry.company.com/my-app:latest

# Extract using credentials
artship cp private-registry.company.com/my-app:latest \
  -a /app/binary \
  -o ./production-binary \
  -u $REGISTRY_USERNAME \
  -p $REGISTRY_PASSWORD

# Or use Docker credentials (recommended)
docker login private-registry.company.com
artship cp private-registry.company.com/my-app:latest \
  -a /app/binary \
  -o ./production-binary
```

### Example 3: Analyzing Third-Party Images

```bash
# Explore a third-party image structure
artship ls redis:latest -d

# Find specific configuration files
artship ls redis:latest -f file | grep -i conf

# Extract Redis configuration for customization
artship cat redis:latest /usr/local/etc/redis/redis.conf > redis-base.conf

# Extract Redis binary for standalone deployment
artship cp redis:latest -a redis-server -o ./redis-server
```

### Command Reference

#### `artship cp`

Copy artifacts from an OCI/Docker image to local filesystem.

**Arguments:**
- `<image>` - OCI/Docker image reference (required)

**Flags:**
- `-a, --artifact` - Artifact names to extract (files or directories, required, can be specified multiple times)  
- `-o, --output` - Target path for the extracted artifact (required, default: current directory)
- `-u, --username` - Username for registry authentication (optional)
- `-p, --password` - Password for registry authentication (optional)
- `-t, --token` - Token for registry authentication (optional)
- `--auth` - Auth string for registry authentication (optional)
- `-k, --insecure` - Allow insecure registry connections (optional)
- `-v, --verbose` - Verbose debug output (optional)
- `-h, --help` - Show help

**Examples:**
```bash
# Extract Redis server binary
artship cp redis:latest -a redis-server -o ./redis-server
Downloading image: redis:latest
Successfully copied 1 artifacts:
  redis-server -> ./redis-server (file, 15.8 MB)
Total size: 15.8 MB
```

```bash
# Extract multiple artifacts
artship cp myapp:latest -a /app/bin -a /app/data -o ./bin -o ./data
Downloading image: myapp:latest
Successfully copied 2 artifacts:
  /app/bin -> ./bin (directory, 0 B)
  /app/data -> ./data (directory, 0 B)
Total size: 2.5 MB
```

#### `artship ls`

List all files and directories available in an OCI/Docker image.

**Arguments:**
- `<image>` - OCI/Docker image reference (required)

**Flags:**
- `-d, --detailed` - Show detailed info (size, type, permissions)
- `-f, --filter` - Filter by type: file, dir, symlink, hardlink, all
- `-u, --username` - Username for registry authentication (optional)
- `-p, --password` - Password for registry authentication (optional)
- `-t, --token` - Token for registry authentication (optional)
- `--auth` - Auth string for registry authentication (optional)
- `-k, --insecure` - Allow insecure registry connections (optional)
- `-v, --verbose` - Verbose debug output (optional)
- `-h, --help` - Show help

**Examples:**
```bash
# List all artifacts
artship ls nginx:latest | head -10
./
bin
boot/
dev/
etc/
etc/.pwd.lock
etc/adduser.conf
```

```bash
# List with detailed information
artship ls nginx:latest --detailed | head -5
TYPE     SIZE       MODE     PATH
-------- ---------- -------- --------
dir      0 B        0755     ./
symlink  -          0777     bin
dir      0 B        0755     boot/
```

```bash
# Filter by file type
artship ls nginx:latest --filter file | grep nginx | head -3
usr/sbin/nginx
usr/sbin/nginx-debug
etc/default/nginx
```

#### `artship cat`

Display the content of a specific file artifact from an OCI/Docker image to stdout.

**Arguments:**
- `<image>` - OCI/Docker image reference (required)
- `<artifact>` - Artifact to show content (required)

**Flags:**
- `-u, --username` - Username for registry authentication (optional)
- `-p, --password` - Password for registry authentication (optional)
- `-t, --token` - Token for registry authentication (optional)
- `--auth` - Auth string for registry authentication (optional)
- `-k, --insecure` - Allow insecure registry connections (optional)
- `-v, --verbose` - Verbose debug output (optional)
- `-h, --help` - Show help

**Examples:**
- `artship cat nginx:latest /etc/nginx/nginx.conf`
- `artship cat alpine:latest /etc/passwd`
- `artship cat private.registry.com/app:latest /config/app.yml -u user -p pass`

#### `artship extract`

Extract all files and directories from an OCI/Docker image to local filesystem.

**Arguments:**
- `<image>` - OCI/Docker image reference (required)

**Flags:**
- `-o, --output` - Target directory to extract all files (default: current directory)
- `-u, --username` - Username for registry authentication (optional)
- `-p, --password` - Password for registry authentication (optional)
- `-t, --token` - Token for registry authentication (optional)
- `--auth` - Auth string for registry authentication (optional)
- `-k, --insecure` - Allow insecure registry connections (optional)
- `-v, --verbose` - Verbose debug output (optional)
- `-h, --help` - Show help

**Examples:**
- `artship extract nginx:latest`
- `artship extract alpine:latest --output ./extracted-alpine`
- `artship extract private.registry.com/app:latest --output ./extracted-app -u user -p pass`

#### `artship has`

Check if a specific artifact exists in an OCI/Docker image.

**Arguments:**
- `<image>` - OCI/Docker image reference (required)
- `<artifact>` - Artifact to check existence (required)

**Flags:**
- `-u, --username` - Username for registry authentication (optional)
- `-p, --password` - Password for registry authentication (optional)
- `-t, --token` - Token for registry authentication (optional)
- `--auth` - Auth string for registry authentication (optional)
- `-k, --insecure` - Allow insecure registry connections (optional)
- `-v, --verbose` - Verbose debug output (optional)
- `-h, --help` - Show help

**Examples:**
- `artship has nginx:latest nginx`
- `artship has nginx:latest /etc/nginx/nginx.conf`
- `artship has private-registry.com/app:latest myapp -u user -p pass`

#### `artship info`

Show detailed information about a specific artifact from an OCI/Docker image.

**Arguments:**
- `<image>` - OCI/Docker image reference (required)
- `<artifact>` - Artifact to show info (required)

**Flags:**
- `-u, --username` - Username for registry authentication (optional)
- `-p, --password` - Password for registry authentication (optional)
- `-t, --token` - Token for registry authentication (optional)
- `--auth` - Auth string for registry authentication (optional)
- `-k, --insecure` - Allow insecure registry connections (optional)
- `-v, --verbose` - Verbose debug output (optional)
- `-h, --help` - Show help

**Examples:**
- `artship info nginx:latest nginx`
- `artship info nginx:latest /etc/nginx/nginx.conf`
- `artship info private-registry.com/app:latest myapp -u user -p pass`

#### `artship meta`

Display detailed metadata information about an OCI/Docker image.

**Arguments:**
- `<image>` - OCI/Docker image reference (required)

**Flags:**
- `-u, --username` - Username for registry authentication (optional)
- `-p, --password` - Password for registry authentication (optional)
- `-t, --token` - Token for registry authentication (optional)
- `--auth` - Auth string for registry authentication (optional)
- `-k, --insecure` - Allow insecure registry connections (optional)
- `-v, --verbose` - Verbose debug output (optional)
- `-h, --help` - Show help

**Examples:**
- `artship meta nginx:latest`
- `artship meta alpine:latest`
- `artship meta ubuntu:20.04`
- `artship meta private.registry.com/app:latest -u user -p pass`

#### `artship tags`

List available tags for an OCI/Docker repository.

**Arguments:**
- `<repository>` - OCI/Docker repository name (required)

**Flags:**
- `-u, --username` - Username for registry authentication (optional)
- `-p, --password` - Password for registry authentication (optional)
- `-t, --token` - Token for registry authentication (optional)
- `--auth` - Auth string for registry authentication (optional)
- `-k, --insecure` - Allow insecure registry connections (optional)
- `-v, --verbose` - Verbose debug output (optional)
- `-h, --help` - Show help

**Examples:**
- `artship tags nginx`
- `artship tags gcr.io/my-project/my-app`
- `artship tags private-registry.com/app -u user -p pass`

#### `artship diff`

Compare filesystems between two OCI/Docker images and show differences.

**Arguments:**
- `<image1>` - First OCI/Docker image reference (required)
- `<image2>` - Second OCI/Docker image reference (required)

**Flags:**
- `-o, --output` - Output format: json (optional, default: colored text)
- `--show-unchanged` - Show unchanged files in output (optional)
- `-f, --filter` - Filter results: added, removed, modified, all (optional)
- `-u, --username` - Username for registry authentication (optional)
- `-p, --password` - Password for registry authentication (optional)
- `-t, --token` - Token for registry authentication (optional)
- `--auth` - Auth string for registry authentication (optional)
- `-k, --insecure` - Allow insecure registry connections (optional)
- `-v, --verbose` - Verbose debug output (optional)
- `-h, --help` - Show help

**Examples:**
```bash
# Compare two image versions
artship diff nginx:1.24 nginx:1.25

# JSON output for automation
artship diff nginx:latest nginx:alpine -o json

# Show only added files
artship diff node:18 node:20 --filter added

# Compare private registry images
artship diff registry.io/app:v1 registry.io/app:v2 -u user -p pass
```

#### `artship mirror`

Copy/mirror an OCI/Docker image from source to destination registry.

**Arguments:**
- `<source-image>` - Source OCI/Docker image reference (required)
- `<destination-image>` - Destination OCI/Docker image reference (required)

**Flags:**
- `-u, --username` - Username for destination registry authentication (optional)
- `-p, --password` - Password for destination registry authentication (optional)
- `-t, --token` - Token for destination registry authentication (optional)
- `--auth` - Auth string for destination registry authentication (optional)
- `--src-username` - Username for source registry (if different from destination) (optional)
- `--src-password` - Password for source registry (if different from destination) (optional)
- `--src-token` - Token for source registry (if different from destination) (optional)
- `--src-auth` - Auth string for source registry (if different from destination) (optional)
- `--src-insecure` - Allow insecure connections to source registry (optional)
- `--dest-insecure` - Allow insecure connections to destination registry (optional)
- `-v, --verbose` - Verbose debug output (optional)
- `-h, --help` - Show help

**Examples:**
```bash
# Copy from Docker Hub to private registry
artship mirror nginx:latest myregistry.com/nginx:latest -u admin -p secret

# Different credentials for source and destination
artship mirror gcr.io/private/app:v1 registry.company.com/app:v1 \
  --src-username _json_key --src-password "$(cat key.json)" \
  -u admin -p secret

# Work with insecure registries
artship mirror insecure.io/app:latest registry.company.com/app:latest \
  --src-insecure -u admin -p secret

# Rename/retag images
artship mirror myregistry.com/app:v1.0 myregistry.com/app:latest -u admin -p secret
```

#### `artship version`

Print version information.

**Examples:**
- `artship version`

## Development

### Prerequisites

- Go 1.25 or later
- Docker (for building container images)
- make

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build multi-platform Docker images
make build-images

# Push Docker image to registry (multi-platform)
make push
```

### Project Structure

```
.
├── cmd/                    # Application entry point
│   └── main.go            # Main function
├── internal/              # Private packages
│   ├── command/           # CLI commands
│   │   ├── root.go       # Root command and main CLI setup
│   │   ├── copy.go       # Copy subcommand (extract artifacts)
│   │   ├── list.go       # List subcommand (browse artifacts with filtering)
│   │   ├── cat.go        # Cat subcommand (display file content)
│   │   ├── extract.go    # Extract subcommand (extract all files)
│   │   ├── has.go        # Has subcommand (check artifact existence)
│   │   ├── info.go       # Info subcommand (detailed artifact info)
│   │   ├── meta.go       # Meta subcommand (image metadata)
│   │   ├── tags.go       # Tags subcommand (list repository tags)
│   │   ├── diff.go       # Diff subcommand (compare images)
│   │   ├── mirror.go     # Mirror subcommand (copy between registries)
│   │   └── version.go    # Version subcommand
│   ├── client/            # Core business logic
│   │   ├── client.go     # Main client with authentication
│   │   ├── copy.go       # Artifact copying functionality
│   │   ├── list.go       # Artifact listing functionality
│   │   ├── cat.go        # File content retrieval
│   │   ├── extract.go    # Full image extraction
│   │   ├── has.go        # Artifact existence checking
│   │   ├── info.go       # Detailed artifact information
│   │   ├── meta.go       # Image metadata retrieval
│   │   ├── tags.go       # Repository tag listing
│   │   ├── diff.go       # Image comparison functionality
│   │   └── mirror.go     # Image mirroring functionality
│   ├── tools/             # Utility functions
│   │   ├── copy.go       # File operations with progress
│   │   ├── walk.go       # Tar archive traversal
│   │   ├── name.go       # Artifact matching logic
│   │   └── format.go     # Data formatting utilities
│   ├── logs/              # Logging functionality
│   │   ├── logger.go     # Logger implementation
│   │   └── colors.go     # Color output support
│   └── version/           # Version information
│       └── version.go    # Version handling and formatting
```

## How It Works

1. **Image Download**: Uses `google/go-containerregistry` to pull OCI/Docker images from registries
2. **Authentication**: Supports username/password, token, auth string authentication or uses Docker's credential keychain
3. **Layer Extraction**: Iterates through all image layers to find the target artifacts
4. **Artifact Matching**: Supports exact path matches, filename matches, and directory content extraction
5. **Multi-Type Support**: Handles regular files, directories, symbolic links, and hard links
6. **Path Resolution**: Automatically resolves relative paths to absolute paths
7. **Directory Creation**: Creates target directories if they don't exist
8. **Size Reporting**: Shows the size of extracted artifacts in human-readable format

## Artifact Matching

Artship supports flexible artifact matching:

- **Exact path match**: `/usr/bin/nginx` matches exactly `/usr/bin/nginx`
- **Filename match**: `nginx` matches any file named `nginx` in any directory
- **Directory extraction**: `/app/bin` extracts all contents of the `/app/bin` directory
- **Multiple artifacts**: Specify multiple `--artifact` and `--output` pairs
- **Size reporting**: Shows extracted file sizes and totals

## Benefits Over Alternative Approaches

- ✅ **No container runtime required** - Works in CI/CD environments without Docker daemon
- ✅ **Faster** - Direct layer access without container startup
- ✅ **Targeted extraction** - No need to export entire filesystem
- ✅ **Stateless operation** - No container lifecycle management
- ✅ **Security** - No container execution, just file extraction
- ✅ **Resource efficient** - Minimal memory and CPU usage

## Troubleshooting

### Artifact not found
If you get an error like "artifacts not found in image":
- Check that the artifact name/path is correct and exists in the image
- Use `artship ls <image>` to explore the image contents first
- Try exact paths when possible: `/usr/bin/nginx` instead of `nginx`

### Permission errors
If you get permission errors when extracting to system paths:
- Run with sudo: `sudo artship cp ...`
- Or copy to a user directory: `artship cp ... -o ~/bin/binary`

### Registry authentication

**Option 1: Using flags**
```bash
artship cp private.registry.com/app:latest -a app -o ./app -u username -p password
```

**Option 2: Using Docker credentials (recommended)**
Artship automatically uses Docker's credential keychain if no username/password are provided:
```bash
docker login private.registry.com
artship cp private.registry.com/app:latest -a app -o ./app
```

**Option 3: Environment variables**
```bash
export DOCKER_CONFIG=/path/to/.docker
artship cp private.registry.com/app:latest -a app -o ./app
```

### Mismatched artifacts and outputs
If you get "number of artifacts must match number of target paths":
- Ensure you have the same number of `--artifact` and `--output` flags
- Each artifact needs a corresponding output path

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [go-containerregistry](https://github.com/google/go-containerregistry) for OCI image handling
- [Cobra](https://github.com/spf13/cobra) for CLI framework
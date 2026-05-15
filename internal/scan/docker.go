package scan

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// DiscoverDocker checks Docker disk usage via `docker system df` and returns
// entries for images, volumes, and build cache. Returns empty slice silently
// if Docker is not installed or the daemon is not running.
func DiscoverDocker(home string) []Entry {
	cmd := exec.Command("docker", "system", "df")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	return parseDockerDF(string(out), home)
}

func parseDockerDF(output, home string) []Entry {
	var entries []Entry
	lines := strings.Split(strings.TrimSpace(output), "\n")

	var imagesSize, volumesSize, buildCacheSize int64

	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "TYPE") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		// Single-word types (Images, Containers) produce 6 fields.
		// Two-word types (Local Volumes, Build Cache) produce 7 fields.
		offset := len(fields) - 6
		if offset < 0 {
			offset = 0
		}

		category := fields[0]
		if offset > 0 {
			category = fields[0] + " " + fields[1]
		}

		reclaimableIdx := 4 + offset
		if reclaimableIdx >= len(fields) {
			continue
		}

		reclaimable := parseDockerSize(fields[reclaimableIdx])

		switch category {
		case "Images":
			imagesSize = reclaimable
		case "Local Volumes":
			volumesSize = reclaimable
		case "Build Cache":
			buildCacheSize = reclaimable
		}
	}

	basePath := filepath.Join(home, "Library/Containers/com.docker.docker/Data")

	entries = append(entries, Entry{
		Path:        basePath,
		Label:       "Docker images",
		Description: "Container images",
		SizeMB:      imagesSize,
		Tier:        TierApp,
		Cleanable:   false,
		CleanHint:   "docker image prune -a",
		StackTag:    "docker",
	})

	entries = append(entries, Entry{
		Path:        basePath,
		Label:       "Docker volumes",
		Description: "Persistent storage volumes",
		SizeMB:      volumesSize,
		Tier:        TierApp,
		Cleanable:   false,
		CleanHint:   "docker volume prune",
		StackTag:    "docker",
	})

	if buildCacheSize > 0 {
		entries = append(entries, Entry{
			Path:        basePath,
			Label:       "Docker build cache",
			Description: "Unused build layers",
			SizeMB:      buildCacheSize,
			Tier:        TierApp,
			Cleanable:   false,
			CleanHint:   "docker builder prune",
			StackTag:    "docker",
		})
	}

	return entries
}

func parseDockerSize(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "0B" {
		return 0
	}
	var value float64
	var unit string
	fmt.Sscanf(s, "%f%s", &value, &unit)
	switch strings.ToUpper(unit) {
	case "GB":
		return int64(value * 1024)
	case "MB":
		return int64(value)
	case "KB":
		return int64(value / 1024)
	default:
		return 0
	}
}

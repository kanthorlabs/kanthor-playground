package kanthorsdk

import (
	"errors"
	"strings"
)

func host(proj *Project, credentials string) (string, error) {
	segments := strings.Split(credentials, ":")
	if len(segments) != 2 || len(segments[1]) == 0 {
		return "", errors.New("malform credentials")
	}

	parts := strings.Split(segments[1], ".")
	if len(parts) == 2 {
		if h, exist := proj.Hosts[parts[0]]; exist {
			return h, nil
		}
	}

	return proj.Host, nil
}

func scheme(host string) string {
	parts := strings.Split(host, ":")
	if strings.HasPrefix(parts[0], "localhost") {
		return "http"
	}
	if strings.HasPrefix(parts[0], "127.0.0.1") {
		return "http"
	}
	// for k8s services
	if strings.HasSuffix(parts[0], ".local") {
		return "http"
	}
	// for docker compose services
	if !strings.Contains(parts[0], ".") {
		return "http"
	}
	return "https"
}

func max[T int | int32 | int64](x, y T) T {
	if x < y {
		return y
	}
	return x
}

func min[T int | int32 | int64](x, y T) T {
	if x > y {
		return y
	}
	return x
}

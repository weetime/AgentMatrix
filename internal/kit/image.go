package kit

import (
	"errors"
	"strings"
)

const (
	defaultDomain string = "docker.io"
)

// artifact 要求为 [<domain>/]<project>/<repository>:<tag>格式
func parseArtifact(artifact string) (project, repository, tag string, err error) {
	tagIndex := strings.LastIndexByte(artifact, ':')
	if tagIndex == -1 {
		err = errors.New("tag is not specified in artifact")
		return
	}
	tag = artifact[tagIndex+1:]
	if !imageTagRegex.MatchString(tag) {
		err = errors.New("invalid image tag format")
		return
	}

	// 自动解析domain
	// 参考 https://github.com/distribution/reference/blob/8c942b0459dfdcc5b6685581dd0a5a470f615bff/normalize.go#L146
	domain, remoteName := defaultDomain, artifact[:tagIndex]
	maybeDomain, maybeRemoteName, ok := strings.Cut(artifact[:tagIndex], "/")
	if ok {
		switch {
		case maybeDomain == "localhost":
			domain, remoteName = maybeDomain, maybeRemoteName
		case maybeDomain == "index.docker.io":
			remoteName = maybeRemoteName
		case strings.ContainsAny(maybeDomain, ".:"):
			domain, remoteName = maybeDomain, maybeRemoteName
		case strings.ToLower(maybeDomain) != maybeDomain:
			domain, remoteName = maybeDomain, maybeRemoteName
		}
	}
	if domain == defaultDomain && !strings.ContainsRune(remoteName, '/') {
		remoteName = "library/" + remoteName
	}

	if !imageNameRegex.MatchString(remoteName) {
		err = errors.New("invalid image name format")
		return
	}

	project, repository, ok = strings.Cut(remoteName, "/")
	if !ok {
		err = errors.New("project or repository is not specified in artifact")
		return
	}
	return
}

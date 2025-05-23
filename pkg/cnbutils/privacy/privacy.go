package privacy

import (
	"crypto/sha256"
	"fmt"
	"strings"

	containerName "github.com/google/go-containerregistry/pkg/name"
)

var allowedBuildpackSources = []struct {
	registry, repositoryPrefix string
}{
	// Paketo
	{
		registry:         "docker.io",
		repositoryPrefix: "paketobuildpacks/",
	}, {
		registry:         "index.docker.io",
		repositoryPrefix: "paketobuildpacks/",
	},
	// Google Buildpacks
	{
		registry:         "gcr.io",
		repositoryPrefix: "buildpacks/",
	},
	// Heroku
	{
		registry:         "public.ecr.aws",
		repositoryPrefix: "heroku-buildpacks/",
	},
}

func FilterBuilder(builder string) string {
	result := FilterBuildpacks([]string{builder})
	return result[0]
}

// FilterBuildpacks filters a list of buildpacks to redact Personally Identifiable Information (PII) like the hostname of a personal registry
func FilterBuildpacks(buildpacks []string) []string {
	result := make([]string, 0, len(buildpacks))
	hash := sha256.New()

	for _, buildpack := range buildpacks {
		ref, err := containerName.ParseReference(strings.ToLower(buildpack))
		if err != nil {
			result = append(result, "<error>")
			continue
		}

		registry := ref.Context().Registry.Name()
		repository := ref.Context().RepositoryStr()

		allowed := false
		for _, allowedBuildpackSource := range allowedBuildpackSources {
			if registry == allowedBuildpackSource.registry && strings.HasPrefix(repository, allowedBuildpackSource.repositoryPrefix) {
				allowed = true
				break
			}
		}

		if allowed {
			result = append(result, buildpack)
		} else {
			hash.Write([]byte(buildpack))
			result = append(result, fmt.Sprintf("%x", hash.Sum(nil)))
			hash.Reset()
		}
	}
	return result
}

var allowedEnvKeys = map[string]interface{}{
	// Java
	// https://github.com/paketo-buildpacks/sap-machine and https://github.com/paketo-buildpacks/bellsoft-liberica
	"BP_JVM_VERSION": nil,
	"BP_JVM_TYPE":    nil,
	// https://github.com/paketo-buildpacks/apache-tomcat
	"BP_TOMCAT_VERSION": nil,

	// Node
	// https://github.com/paketo-buildpacks/node-engine
	"BP_NODE_VERSION": nil,
}

// FilterEnv filters a map of environment variables to redact Personally Identifiable Information (PII)
func FilterEnv(in map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for key, value := range in {
		_, allowed := allowedEnvKeys[key]
		if allowed {
			out[key] = value
		}
	}
	return out
}

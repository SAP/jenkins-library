package format

import "github.com/anchore/packageurl-go"

func TransformBuildToPurlType(buildType string) string {
	switch buildType {
	case "maven":
		return packageurl.TypeMaven
	case "npm":
		return packageurl.TypeNPM
	case "docker":
		return packageurl.TypeDocker
	case "kaniko":
		return packageurl.TypeDocker
	case "golang":
		return packageurl.TypeGolang
	case "mta":
		return packageurl.TypeComposer
	}
	return packageurl.TypeGeneric
}

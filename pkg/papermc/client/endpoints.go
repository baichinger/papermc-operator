package client

import (
	"fmt"
)

const (
	paperApiUrl           = "https://api.papermc.io"
	paperVersionEndpoint  = "/v2/projects/paper/versions/%s"
	paperBuildEndpoint    = "/v2/projects/paper/versions/%s/builds/%d"
	paperDownloadEndpoint = "/v2/projects/paper/versions/%s/builds/%d/downloads/%s"
)

func getUrlForVersionDetails(version string) string {
	endpoint := fmt.Sprintf(paperVersionEndpoint, version)
	return fmt.Sprintf("%s%s", paperApiUrl, endpoint)
}

func getUrlForVersionBuildDetails(version string, build int) string {
	endpoint := fmt.Sprintf(paperBuildEndpoint, version, build)
	return fmt.Sprintf("%s%s", paperApiUrl, endpoint)
}

func getUrlForVersionBuildArtifactDownload(version string, build int, artifact string) string {
	endpoint := fmt.Sprintf(paperDownloadEndpoint, version, build, artifact)
	return fmt.Sprintf("%s%s", paperApiUrl, endpoint)
}

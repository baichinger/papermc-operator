package client

type Client interface {
	GetBuildForVersion(version string) (int, error)
	GetUrlForVersionBuildDownload(version string, build int) (string, error)
}

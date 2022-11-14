package client

// Client is the client for interacting with the Paper MC API. Only minimal functionality is provided for downloading
// new versions of Paper.
type Client interface {
	// GetBuildForVersion determines the build id for a given version.
	GetBuildForVersion(version string) (int, error)

	// GetUrlForVersionBuildDownload determines the download URL of a given version/build. The URL returned points
	// to the corresponding JAR file.
	GetUrlForVersionBuildDownload(version string, build int) (string, error)
}

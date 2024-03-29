package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func NewPapermcClient(ctx context.Context) Client {
	return &papermcClient{
		Client: http.DefaultClient,
		Logger: log.FromContext(ctx),
	}
}

type papermcClient struct {
	*http.Client
	logr.Logger
}

func (c *papermcClient) GetBuildForVersion(version string) (int, error) {
	response := struct {
		Builds []int `json:"builds"`
	}{}

	err := c.doRequestAndUnmarshal(buildVersionDetailsUrl(version), &response)
	if err != nil {
		return 0, err
	}

	if len(response.Builds) == 0 {
		return 0, fmt.Errorf("no build found")
	}

	return response.Builds[len(response.Builds)-1], nil
}

func (c *papermcClient) GetUrlForVersionBuildDownload(version string, build int) (string, error) {
	artifact, err := c.getArtifactNameForVersionAndBuild(version, build)
	if err != nil {
		return "", err
	}

	return buildVersionBuildArtifactDownloadUrl(version, build, artifact), nil
}

func (c *papermcClient) getArtifactNameForVersionAndBuild(version string, build int) (string, error) {
	response := struct {
		Downloads struct {
			Application struct {
				Name   string `json:"name"`
				Sha265 string `json:"sha265"`
			} `json:"application"`
		} `json:"downloads"`
	}{}

	err := c.doRequestAndUnmarshal(buildVersionBuildDetailsUrl(version, build), &response)
	if err != nil {
		return "", err
	}

	return response.Downloads.Application.Name, nil
}

func (c *papermcClient) doRequestAndUnmarshal(url string, structuredResponse interface{}) error {
	response, err := c.doRequest(url)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()

	responseData, err := getResponseData(response)
	if err != nil {
		return err
	}

	return json.Unmarshal(responseData, &structuredResponse)
}

func (c *papermcClient) doRequest(url string) (*http.Response, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize http request: %s", err)
	}

	c.Logger.V(2).Info("PaperMC API request", "url", request.URL.String())

	return c.Client.Do(request)
}

func getResponseData(response *http.Response) ([]byte, error) {
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("paper API returned invalid status code: %d", response.StatusCode)
	}

	return io.ReadAll(response.Body)
}

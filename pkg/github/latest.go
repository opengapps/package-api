package github

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nezorflame/opengapps-mirror-bot/pkg/gapps"
)

const latestReleaseURLTemplate = "https://raw.githubusercontent.com/opengapps/%s/master/LATEST.json"

var releaseURLMap = map[gapps.Platform]string{
	gapps.PlatformArm:    fmt.Sprintf(latestReleaseURLTemplate, gapps.PlatformArm),
	gapps.PlatformArm64:  fmt.Sprintf(latestReleaseURLTemplate, gapps.PlatformArm64),
	gapps.PlatformX86:    fmt.Sprintf(latestReleaseURLTemplate, gapps.PlatformX86),
	gapps.PlatformX86_64: fmt.Sprintf(latestReleaseURLTemplate, gapps.PlatformX86_64),
}

// LatestRelease describes the latest gapps release
type LatestRelease struct {
	Arch   string         `json:"arch"`
	Date   string         `json:"date"`
	Assets []ReleaseAsset `json:"assets"`
}

// ReleaseAsset describes the gapps release for API and its available variants
type ReleaseAsset struct {
	API      string   `json:"api"`
	Variants []string `json:"variants"`
}

// GetLatestRelease returns the latest release info for the selected architecture
func (c *client) GetLatestRelease(ctx context.Context, arch gapps.Platform) (*LatestRelease, error) {
	req, err := c.client.NewRequest(http.MethodGet, releaseURLMap[arch], nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request for the LATEST file for arch '%s': %w", arch, err)
	}

	var release LatestRelease
	resp, err := c.client.Do(ctx, req, &release)
	if err != nil {
		return nil, fmt.Errorf("unable to acquire LATEST file for arch '%s': %w", arch, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to acquire LATEST file for arch '%s': got response '%s'", arch, resp.Status)
	}

	return &release, nil
}

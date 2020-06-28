package checks

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/flanksource/canary-checker/pkg"
)

var (
	dockerClient *client.Client

	size = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "canary_check_image_size",
			Help: "Size of docker image",
		},
		[]string{"image"},
	)

	imagePullTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "canary_check_image_pull_time",
			Help:    "Image pull time",
			Buckets: []float64{100, 500, 1000, 5000, 15000, 30000},
		},
		[]string{"image"},
	)
)

func init() {
	var err error
	dockerClient, err = client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	prometheus.MustRegister(size, imagePullTime)
}

type DockerPullChecker struct{}

func (c *DockerPullChecker) Run(config pkg.Config, results chan *pkg.CheckResult) {
	for _, conf := range config.DockerPull {
		results <- c.Check(conf.DockerPullCheck)
	}
}

// Type: returns checker type
func (c *DockerPullChecker) Type() string {
	return "docker-pull"
}

// Run: Check every entry from config according to Checker interface
// Returns check result and metrics
func (c *DockerPullChecker) Check(check pkg.DockerPullCheck) *pkg.CheckResult {
	start := time.Now()
	ctx := context.Background()
	digestVerified, sizeVerified := false, false
	authConfig := types.AuthConfig{
		Username: check.Username,
		Password: check.Password,
	}
	encodedJSON, _ := json.Marshal(authConfig)
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	out, err := dockerClient.ImagePull(ctx, check.Image, types.ImagePullOptions{RegistryAuth: authStr})
	elapsed := time.Since(start)
	if err != nil {
		return Failf(check, "Failed to pull image: %s", err)
	} else {
		buf := new(bytes.Buffer)
		defer out.Close()
		_, _ = buf.ReadFrom(out)
		if strings.Contains(buf.String(), check.ExpectedDigest) {
			digestVerified = true
		}
	}

	inspect, _, _ := dockerClient.ImageInspectWithRaw(ctx, check.Image)
	if inspect.Size == check.ExpectedSize {
		sizeVerified = true
	}
	m := []pkg.Metric{
		{
			Name: "pull_time", Type: pkg.HistogramType,
			Labels: map[string]string{"image": check.Image},
			Value:  float64(elapsed.Milliseconds()),
		},
		{
			Name: "totalLayers", Type: pkg.GaugeType,
			Labels: map[string]string{"image": check.Image},
			Value:  float64(len(inspect.RootFS.Layers)),
		},
		{
			Name: "size", Type: pkg.GaugeType,
			Labels: map[string]string{"image": check.Image},
			Value:  float64(inspect.Size),
		},
	}

	size.WithLabelValues(check.Image).Set(float64(inspect.Size))
	imagePullTime.WithLabelValues(check.Image).Observe(float64(elapsed.Milliseconds()))
	return &pkg.CheckResult{
		Check:    check,
		Pass:     digestVerified && sizeVerified,
		Invalid:  !(digestVerified && sizeVerified),
		Duration: elapsed.Milliseconds(),
		Metrics:  m,
	}
}

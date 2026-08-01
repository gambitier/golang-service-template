package platform

import (
	"strings"

	commonobservability "github.com/gambitier/go-pkgs/observability"
)

// ObservabilityFromYAML normalizes the opentel YAML block into a runtime Config.
// Kept here so go-pkgs/observability stays independent of any app config loader.
func ObservabilityFromYAML(fallbackServiceName string, yamlCfg commonobservability.YAMLConfig) commonobservability.Config {
	serviceName := strings.TrimSpace(yamlCfg.ServiceName)
	if serviceName == "" {
		serviceName = strings.TrimSpace(fallbackServiceName)
	}

	insecure := yamlCfg.Insecure
	if strings.TrimSpace(yamlCfg.InsecureMode) != "" {
		switch strings.TrimSpace(strings.ToLower(yamlCfg.InsecureMode)) {
		case "1", "true", "yes", "on":
			insecure = true
		case "0", "false", "no", "off":
			insecure = false
		}
	}

	ratio := yamlCfg.Sampling.Ratio
	if ratio <= 0 || ratio > 1 {
		ratio = 1.0
	}

	headers := yamlCfg.Headers
	if headers == nil {
		headers = map[string]string{}
	}

	return commonobservability.Config{
		ServiceName:           serviceName,
		Enabled:               yamlCfg.Enabled,
		Endpoint:              strings.TrimSpace(yamlCfg.CollectorURL),
		Headers:               headers,
		Insecure:              insecure,
		TracesEndpoint:        strings.TrimSpace(yamlCfg.TracesURL),
		MetricsEndpoint:       strings.TrimSpace(yamlCfg.MetricsURL),
		LogsEndpoint:          strings.TrimSpace(yamlCfg.LogsURL),
		SamplingRatio:         ratio,
		DeploymentEnvironment: strings.TrimSpace(yamlCfg.Resource.DeploymentEnvironment),
		ServiceVersion:        strings.TrimSpace(yamlCfg.Resource.ServiceVersion),
	}
}

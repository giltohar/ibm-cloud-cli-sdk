package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/bluemix/authentication"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/bluemix/configuration/core_config"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/bluemix/consts"
	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/common/rest"
)

type pluginContext struct {
	core_config.ReadWriter
	cfConfig     cfConfigWrapper
	pluginConfig PluginConfig
	pluginPath   string
}

type cfConfigWrapper struct {
	core_config.CFConfig
}

func (c cfConfigWrapper) RefreshUAAToken() (string, error) {
	if !c.HasAPIEndpoint() {
		return "", fmt.Errorf("CloudFoundry API endpoint is not set")
	}

	config := &authentication.UAAConfig{UAAEndpoint: c.AuthenticationEndpoint()}
	auth := authentication.NewUAARepository(config, rest.NewClient())
	token, err := auth.RefreshToken(c.UAARefreshToken())
	if err != nil {
		return "", err
	}

	c.SetUAAToken(token.Token())
	c.SetUAARefreshToken(token.RefreshToken)
	return token.Token(), nil
}

func createPluginContext(pluginPath string, coreConfig core_config.ReadWriter) *pluginContext {
	return &pluginContext{
		pluginPath:   pluginPath,
		pluginConfig: loadPluginConfigFromPath(filepath.Join(pluginPath, "config.json")),
		ReadWriter:   coreConfig,
		cfConfig:     cfConfigWrapper{coreConfig.CFConfig()},
	}
}

func (c *pluginContext) APIEndpoint() string {
	if compareVersion(c.SDKVersion(), "0.1.1") < 0 {
		return c.ReadWriter.CFConfig().APIEndpoint()
	}
	return c.ReadWriter.APIEndpoint()
}

func compareVersion(v1, v2 string) int {
	s1 := strings.Split(v1, ".")
	s2 := strings.Split(v2, ".")

	n := len(s1)
	if len(s2) > n {
		n = len(s2)
	}

	for i := 0; i < n; i++ {
		var p1, p2 int
		if len(s1) > i {
			p1, _ = strconv.Atoi(s1[i])
		}
		if len(s2) > i {
			p2, _ = strconv.Atoi(s2[i])
		}
		if p1 > p2 {
			return 1
		}
		if p1 < p2 {
			return -1
		}
	}
	return 0
}

func (c *pluginContext) HasAPIEndpoint() bool {
	return c.APIEndpoint() != ""
}

func (c *pluginContext) PluginDirectory() string {
	return c.pluginPath
}

func (c *pluginContext) PluginConfig() PluginConfig {
	return c.pluginConfig
}

func (c *pluginContext) RefreshIAMToken() (string, error) {
	endpoint := os.Getenv("IAM_ENDPOINT")
	if endpoint == "" {
		endpoint = c.IAMEndpoint()
	}
	if endpoint == "" {
		return "", fmt.Errorf("IAM endpoint is not set")
	}

	config := &authentication.IAMConfig{TokenEndpoint: endpoint + "/identity/token"}
	auth := authentication.NewIAMAuthRepository(config, rest.NewClient())
	iamToken, err := auth.RefreshToken(c.IAMRefreshToken())
	if err != nil {
		return "", err
	}

	c.SetIAMToken(iamToken.Token())
	c.SetIAMRefreshToken(iamToken.RefreshToken)

	return iamToken.Token(), nil
}

func (c *pluginContext) Trace() string {
	return getFromEnvOrConfig(consts.ENV_BLUEMIX_TRACE, c.ReadWriter.Trace())
}

func (c *pluginContext) ColorEnabled() string {
	return getFromEnvOrConfig(consts.ENV_BLUEMIX_COLOR, c.ReadWriter.ColorEnabled())
}

func (c *pluginContext) VersionCheckEnabled() bool {
	return !c.CheckCLIVersionDisabled()
}

func (c *pluginContext) CF() CFContext {
	return c.cfConfig
}

func (c *pluginContext) HasTargetedCF() bool {
	return c.cfConfig.HasAPIEndpoint()
}

func getFromEnvOrConfig(envKey string, config string) string {
	if envVal := os.Getenv(envKey); envVal != "" {
		return envVal
	}
	return config
}

func (c *pluginContext) CommandNamespace() string {
	return os.Getenv(consts.ENV_BLUEMIX_PLUGIN_NAMESPACE)
}

func (c *pluginContext) CLIName() string {
	cliName := os.Getenv(consts.ENV_BLUEMIX_CLI)
	if cliName == "" {
		cliName = "bx"
	}
	return cliName
}

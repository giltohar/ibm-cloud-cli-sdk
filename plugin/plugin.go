// Package plugin provides types and functions common among plugins.
//
// See examples in "github.com/IBM-Cloud/ibm-cloud-cli-sdk/plugin_examples".
package plugin

import (
	"fmt"
	"strings"

	"github.com/IBM-Cloud/ibm-cloud-cli-sdk/bluemix/models"
)

// PluginMetadata describes metadata of a plugin.
type PluginMetadata struct {
	Name          string      // name of the plugin
	Version       VersionType // version of the plugin
	MinCliVersion VersionType // minimal version of CLI required by the plugin
	Namespaces    []Namespace // list of namespaces provided by the plugin
	Commands      []Command   // list of commands provided by the plugin

	// SDKVersion is SDK version used by the plugin.
	// It is set by the plugin framework to check SDK compatibility with the CLI.
	SDKVersion VersionType
}

// VersionType describes the version info
type VersionType struct {
	Major int // major version
	Minor int // minimal version
	Build int // build number
}

func (v VersionType) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Build)
}

// Namespace represents a category of commands that have similar
// functionalities. A command under a namespace is run using 'bx [namespace]
// [command]'.
//
// Some namespaces are predefined by Bluemix CLI and sharable among plugins;
// others are non-shared and defined in each plugin. The Plugin can reference a
// predefined namespace or define a non-shared namespace of its own.
//
// Namespace also supports hierarchy. For example, namespace 'A' can have a sub
// namespace 'B'. And the full qualified name of namespace B is 'A B'.
type Namespace struct {
	Name        string // full qualified name of the namespace
	Description string // description of the namespace
}

// ParentName returns the name of its parent namespace
func (n Namespace) ParentName() string {
	i := strings.LastIndex(n.Name, " ")
	if i < 0 {
		return ""
	}
	return n.Name[:i]
}

// Command describes the metadata of a plugin command
type Command struct {
	Namespace   string // full qualified name of the command's namespace
	Name        string // name of the command
	Alias       string // command alias, usually the command's short name
	Description string // short description of the command
	Usage       string // usage detail to be displayed in command help
	Flags       []Flag // command options
	Hidden      bool   // true to hide the command in help text
}

// FullName returns Command's fully-qualified name prefixed with namespace
func (c Command) FullName() string {
	return strings.TrimSpace(strings.Join([]string{c.Namespace, c.Name}, " "))
}

// FullNames returns Commands's full-qualified names prefixed with namespace
func (c Command) FullNames() []string {
	names := []string{c.FullName()}
	if c.Alias != "" {
		names = append(names, strings.TrimSpace(strings.Join([]string{c.Namespace, c.Alias}, " ")))
	}
	return names
}

// Flag describes a command option
type Flag struct {
	Name        string // name of the option
	Description string // description of the option
	HasValue    bool   // whether the option requires a value or not
	Hidden      bool   // true to hide the option in command help
}

// Plugin is an interface for Bluemix CLI plugins.
type Plugin interface {
	// GetMetadata returns the metadata of the plugin.
	GetMetadata() PluginMetadata

	// Run runs the plugin with given plugin context and arguments.
	//
	// Note: the first arg in args is a command or alias no matter the command
	// has namespace or not. To get the namespace, call
	// PluginContext.CommandNamespace.
	Run(c PluginContext, args []string)
}

// PluginContext is a Bluemix CLI context passed to plugin's Run method. It
// carries service endpoints info, login session, user configuration, plugin
// configuration and provides utility methods.
type PluginContext interface {
	// APIEndpoint returns the targeted API endpoint of IBM Cloud
	APIEndpoint() string

	// HasAPIEndpoint() returns whether an IBM Cloud has been targeted, does not
	// necessarily mean the user has targeted a CF environment.
	// Call HasTargetedCF() to return whether a CF environment has been targeted.
	HasAPIEndpoint() bool

	// ConsoleEndpoint returns the Bluemix Console endpoint
	ConsoleEndpoint() string

	// IAMTEndpoint return the endpoint of IAM token service
	IAMEndpoint() string

	// CloudName returns the name of the target cloud
	CloudName() string

	// CloudType returns the type of the target cloud (like 'public',
	// 'dedicated' etc)
	CloudType() string

	// Region returns the targeted region
	CurrentRegion() models.Region

	// IAMToken returns the IAM access token
	IAMToken() string

	// IAMRefreshToken returns the IAM refresh token
	IAMRefreshToken() string

	// RefreshIAMToken refreshes and returns the IAM access token
	RefreshIAMToken() (string, error)

	// UserEmail returns the Email of the logged in user
	UserEmail() string

	// IsLoggedIn returns if a user has logged into IBM cloud, does not
	// necessarily mean the user has logged in a CF environment.
	// Call CF().IsLoggedIn() to return whether user has been logged into the CF environment.
	IsLoggedIn() bool

	// IMSAccountID returns ID of the IMS account linked to the targeted BSS
	// account
	IMSAccountID() string

	// Account returns the targeted a BSS account
	CurrentAccount() models.Account

	// HasTargetedAccount returns whether an account has been targeted
	HasTargetedAccount() bool

	// ResourceGroup returns the targeted resource group
	CurrentResourceGroup() models.ResourceGroup

	// HasTargetedResourceGroup returns whether a resource group has been targeted
	HasTargetedResourceGroup() bool

	// CF returns the context of the targeted CloudFoundry environment
	CF() CFContext

	// HasTargetedCF returns whether a CloudFoundry environment has been targeted
	HasTargetedCF() bool

	// Locale returns user specified locale
	Locale() string

	// Trace returns user specified trace setting.
	// The value is "true", "false" or path of the trace output file.
	Trace() string

	// ColorEnabled returns whether terminal displays color or not
	ColorEnabled() string

	// IsSSLDisabled returns whether skipping SSL validation or not
	IsSSLDisabled() bool

	// PluginDirectory returns the installation directory of the plugin
	PluginDirectory() string

	// HTTPTimeout returns a timeout for HTTP Client
	HTTPTimeout() int

	// VersionCheckEnabled() returns whether checking for update is performmed
	VersionCheckEnabled() bool

	// PluginConfig returns the plugin specific configuarion
	PluginConfig() PluginConfig

	// CommandNamespace returns the name of the parsed namespace
	CommandNamespace() string

	// CLIName returns binary name of the Bluemix CLI that is invoking the plugin
	CLIName() string
}

// CFContext is a context of the targeted CloudFoundry environment into plugin
type CFContext interface {
	// APIVersion returns the cloud controller API version
	APIVersion() string

	// APIEndpoint returns the Cloud Foundry API endpoint
	APIEndpoint() string

	// HasAPIEndpoint returns whether a Cloud Foundry API endpoint is set
	HasAPIEndpoint() bool

	//DopplerEndpoint returns the Doppler endpoint
	DopplerEndpoint() string

	// UAAEndpoint returns endpoint of UAA token service
	UAAEndpoint() string

	// ISLoggedIn returns if a user has logged into the cloudfoundry instance
	IsLoggedIn() bool

	// Username returns the name of the logged in user
	Username() string

	// UserEmail returns the Email of the logged in user
	UserEmail() string

	// UserGUID returns the GUID of the logged in user
	UserGUID() string

	// UAAToken returns the UAA access token
	// If the token is outdated, call RefreshUAAToken to refresh it
	UAAToken() string

	// UAARefreshToken return the UAA refreshed token
	UAARefreshToken() string

	// RefreshUAAToken refreshes and returns the UAA access token
	RefreshUAAToken() (string, error)

	// CurrentOrganization returns the targeted organization
	CurrentOrganization() models.OrganizationFields

	// HasTargetedOrganization returns if an organization has been targeted
	HasTargetedOrganization() bool

	// CurrentSpace returns the targeted space
	CurrentSpace() models.SpaceFields

	// HasTargetedSpace returns if a space has been targeted
	HasTargetedSpace() bool
}

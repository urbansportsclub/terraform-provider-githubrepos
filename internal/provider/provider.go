package provider

import (
	"context"
	"net/http"
	"os"

	"github.com/google/go-github/v53/github"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var descriptions map[string]string

func init() {
	descriptions = map[string]string{
		"token": "The OAuth token used to connect to GitHub. Anonymous mode is enabled if both `token` and " +
			"`app_auth` are not set.",

		"owner": "The GitHub owner name to manage. " +
			"Use this field instead of `organization` when managing individual accounts.",
	}
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &githubreposProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &githubreposProvider{
			version: version,
		}
	}
}

// githubreposProvider is the provider implementation.
type githubreposProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// githubreposProviderModel maps provider schema data to a Go type.
type githubreposProviderModel struct {
	Token types.String `tfsdk:"token"`
	Owner types.String `tfsdk:"owner"`
}

// Metadata returns the provider type name.
func (p *githubreposProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "githubrepos"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *githubreposProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	// Schema defines the provider-level schema for configuration data.
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Optional:    true,
				Description: descriptions["token"],
			},
			"owner": schema.StringAttribute{
				Optional:    true,
				Description: descriptions["owner"],
			},
		},
	}
}

// Configure prepares a GitHub API client for data sources and resources.
func (p *githubreposProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config githubreposProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown GitHub API Token",
			"The provider cannot create the GitHub API client as there is an unknown configuration value for the GitHub token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the GITHUB_TOKEN environment variable.",
		)
	}

	if config.Owner.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("owner"),
			"Unknown GitHub Organization",
			"The provider cannot create the GitHub API client as there is an unknown configuration value for the GitHub organization name. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the GITHUB_OWNER environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	token := os.Getenv("GITHUB_TOKEN")
	owner := os.Getenv("GITHUB_OWNER")

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	if !config.Owner.IsNull() {
		owner = config.Owner.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing GitHub Token",
			"The provider cannot create the GitHub API client as there is a missing or empty value for the GitHub token. "+
				"Set the token value in the configuration or use the GITHUB_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if owner == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("owner"),
			"Missing GitHub Organization",
			"The provider cannot create the GitHub API client as there is a missing or empty value for the GitHub organization name. "+
				"Set the owner value in the configuration or use the GITHUB_OWNER environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "github_token", token)
	ctx = tflog.SetField(ctx, "github_owner", owner)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "github_token")

	tflog.Debug(ctx, "Creating GitHub client")

	// Create a new GitHub client using the configuration values
	client, err := github.NewEnterpriseClient("https://api.github.com/", "", &http.Client{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create GitHub API Client",
			"An unexpected error occurred when creating the GitHub API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"GitHub Client Error: "+err.Error(),
		)
		return
	}

	// Make the GitHub client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured GitHub client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *githubreposProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// Resources defines the resources implemented in the provider.
func (p *githubreposProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}

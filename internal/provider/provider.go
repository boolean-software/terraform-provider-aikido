// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/boolean-software/aikido-http-client/aikido"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &AikidoProvider{}
var _ provider.ProviderWithFunctions = &AikidoProvider{}

// AikidoProvider defines the provider implementation.
type AikidoProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// AikidoProviderModel describes the provider data model.
type AikidoProviderModel struct {
	ClientId     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

func (p *AikidoProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "aikido"
	resp.Version = p.version
}

func (p *AikidoProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				MarkdownDescription: "Client ID for http api integration",
				Optional:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Client Secret for http api integration",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *AikidoProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data AikidoProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.ClientId.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Unknown Aikido API Client Id",
			"The provider cannot create the Aikido API client as there is an unknown configuration value for the Aikido API client id. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the AIKIDO_CLIENT_ID environment variable.",
		)
	}

	if data.ClientSecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Unknown Aikido API Client Secret",
			"The provider cannot create the Aikido API client as there is an unknown configuration value for the Aikido API client secret. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the AIKIDO_CLIENT_SECRET environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	clientId := os.Getenv("AIKIDO_CLIENT_ID")
	clientSecret := os.Getenv("AIKIDO_CLIENT_SECRET")

	if !data.ClientId.IsNull() {
		clientId = data.ClientId.ValueString()
	}

	if !data.ClientSecret.IsNull() {
		clientSecret = data.ClientSecret.ValueString()
	}

	if clientId == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_id"),
			"Missing HashiCups API Client Id",
			"The provider cannot create the Aikido API client as there is a missing or empty value for the Aikido API client id. "+
				"Set the client id value in the configuration or use the AIKIDO_CLIENT_ID environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if clientSecret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("client_secret"),
			"Missing HashiCups API Client Secret",
			"The provider cannot create the Aikido API client as there is a missing or empty value for the Aikido API client secret. "+
				"Set the client secret value in the configuration or use the AIKIDO_CLIENT_SECRET environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client := aikido.NewAikidoHttpClient(clientId, clientSecret)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *AikidoProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

func (p *AikidoProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
	}
}

func (p *AikidoProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewExampleFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AikidoProvider{
			version: version,
		}
	}
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/boolean-software/aikido-http-client/aikido"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSource = &UsersDataSource{}
var _ datasource.DataSourceWithConfigure = &UsersDataSource{}

func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

type UsersDataSource struct {
	client *aikido.Client
}

type UsersDataSourceModel struct {
	Users []User `tfsdk:"users"`
}

type User struct {
	ID                 int    `tfsdk:"id"`
	FullName           string `tfsdk:"full_name"`
	Email              string `tfsdk:"email"`
	Active             int    `tfsdk:"active"`
	LastLoginTimestamp int    `tfsdk:"last_login_timestamp"`
	Role               string `tfsdk:"role"`
	AuthType           string `tfsdk:"auth_type"`
}

func (d *UsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*aikido.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *aikido.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *UsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int32Attribute{
							MarkdownDescription: "ID of the user",
							Computed:            true,
						},
						"full_name": schema.StringAttribute{
							MarkdownDescription: "Full name of the user",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "Email address of the user",
							Computed:            true,
						},
						"active": schema.Int32Attribute{
							MarkdownDescription: "Status indicating if the user is active",
							Computed:            true,
						},
						"last_login_timestamp": schema.Int32Attribute{
							MarkdownDescription: "Timestamp of the user's last login",
							Computed:            true,
						},
						"role": schema.StringAttribute{
							MarkdownDescription: "Role of the user",
							Computed:            true,
						},
						"auth_type": schema.StringAttribute{
							MarkdownDescription: "Authentication type of the user",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state UsersDataSourceModel

	users, err := d.client.ListUsers(aikido.DefaultListUsersFilters)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Aikido Users",
			err.Error(),
		)
		return
	}

	for _, user := range users {
		stateUser := User{
			ID:                 user.ID,
			FullName:           user.FullName,
			Email:              user.Email,
			Active:             user.Active,
			LastLoginTimestamp: user.LastLoginTimestamp,
			Role:               user.Role,
			AuthType:           user.AuthType,
		}

		state.Users = append(state.Users, stateUser)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

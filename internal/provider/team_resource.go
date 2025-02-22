package provider

import (
	"context"
	"fmt"

	"github.com/boolean-software/aikido-http-client/aikido"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &teamResource{}
	_ resource.ResourceWithConfigure = &teamResource{}
)

func NewTeamResource() resource.Resource {
	return &teamResource{}
}

type teamResource struct {
	client *aikido.Client
}

type teamResourceModel struct {
	ID               types.Int32          `tfsdk:"id"`
	Name             types.String         `tfsdk:"name"`
	Responsibilities []teamResponsibility `tfsdk:"responsibilities"`
}

type teamResponsibility struct {
	ID   types.Int32  `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}

func (r *teamResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (r *teamResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int32Attribute{
				MarkdownDescription: "ID of the team",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int32{int32planmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the team",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"responsibilities": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int32Attribute{
							MarkdownDescription: "ID of resource team is respnsible for",
							Required:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of resource team is responsible for",
							Required:            true,
							Validators: []validator.String{stringvalidator.OneOf(
								"code_repository",
								"container_repository",
								"cloud",
								"domain",
							)},
						},
					},
				},
			},
		},
	}
}

func (r *teamResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *teamResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teamResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTeamRequest := aikido.CreateTeamRequest{
		Name: plan.Name.ValueString(),
	}

	id, err := r.client.CreateTeam(createTeamRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating team",
			"Could not create team, unexpected error: "+err.Error(),
		)
		return
	}

	var responsibilities []aikido.Responsibility

	for _, responsibility := range plan.Responsibilities {
		responsibilities = append(responsibilities, aikido.Responsibility{
			ID:   int(responsibility.ID.ValueInt32()),
			Type: responsibility.Type.ValueString(),
		})
	}

	updateTeamRequest := aikido.UpdateTeamRequest{
		ID:               id,
		Name:             createTeamRequest.Name,
		Responsibilities: responsibilities,
	}

	_, err = r.client.UpdateTeam(updateTeamRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating team",
			"Could not update team, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.Int32Value(updateTeamRequest.ID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *teamResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teamResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	remoteTeam, teamFound, err := r.findTeam(int(state.ID.ValueInt32()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Team",
			"Could not find teams from aikido: "+err.Error(),
		)
		return
	}

	if !teamFound {
		resp.Diagnostics.AddError(
			"Error Reading Team",
			"Could not find team "+state.ID.String()+"|"+state.Name.String()+" from aikido. State might need to be cleaned.",
		)
		return
	}
	state.ID = types.Int32Value(int32(remoteTeam.ID))
	state.Name = types.StringValue(remoteTeam.Name)
	state.Responsibilities = []teamResponsibility{}
	for _, responsibility := range remoteTeam.Responsibilities {
		state.Responsibilities = append(state.Responsibilities, teamResponsibility{
			ID:   types.Int32Value(int32(responsibility.ID)),
			Type: types.StringValue(responsibility.Type),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *teamResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan teamResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateRequest := *plan.toUpdateTeamRequest()

	_, err := r.client.UpdateTeam(updateRequest)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating team",
			"Could not update team, unexpected error: "+err.Error(),
		)
		return
	}

	remoteTeam, teamFound, err := r.findTeam(int(plan.ID.ValueInt32()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Team",
			"Could not find team from aikido: "+err.Error(),
		)
		return
	}

	if teamFound {
		plan.ID = types.Int32Value(int32(remoteTeam.ID))
		plan.Name = types.StringValue(remoteTeam.Name)
		plan.Responsibilities = []teamResponsibility{}
		for _, responsibility := range remoteTeam.Responsibilities {
			plan.Responsibilities = append(plan.Responsibilities, teamResponsibility{
				ID:   types.Int32Value(int32(responsibility.ID)),
				Type: types.StringValue(responsibility.Type),
			})
		}

		diags = resp.State.Set(ctx, &plan)
		resp.Diagnostics.Append(diags...)
	}
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *teamResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teamResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeleteTeam(aikido.DeleteTeamRequest{ID: state.ID.ValueInt32()})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Team",
			"Could not delete team, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *teamResource) fetchTeamsAtPage(page int) ([]aikido.Team, error) {
	teams, err := r.client.ListTeams(aikido.ListTeamsFilters{Page: int32(page), PerPage: 10})
	if err != nil {
		return []aikido.Team{}, err
	}
	return teams, nil
}

func (r *teamResource) findTeam(id int) (*aikido.Team, bool, error) {
	var remoteTeam *aikido.Team

	page := 0
	for {
		teams, err := r.fetchTeamsAtPage(page)
		if err != nil {
			return nil, false, err
		}

		for _, team := range teams {
			if id == team.ID {
				remoteTeam = &team
				page = -1
				break
			}
		}

		if page == -1 || len(teams) == 0 {
			break
		} else {
			page = page + 1
		}
	}

	return remoteTeam, remoteTeam != nil, nil
}

func (t teamResourceModel) toUpdateTeamRequest() *aikido.UpdateTeamRequest {
	var responsibilities []aikido.Responsibility

	for _, responsibility := range t.Responsibilities {
		responsibilities = append(responsibilities, aikido.Responsibility{
			ID:   int(responsibility.ID.ValueInt32()),
			Type: responsibility.Type.ValueString(),
		})
	}

	return &aikido.UpdateTeamRequest{
		ID:               t.ID.ValueInt32(),
		Name:             t.Name.ValueString(),
		Responsibilities: responsibilities,
	}
}

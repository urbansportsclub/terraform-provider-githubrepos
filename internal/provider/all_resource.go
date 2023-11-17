package provider

import (
	"context"
	"fmt"

	"github.com/google/go-github/v53/github"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &allResource{}
	_ resource.ResourceWithConfigure = &allResource{}
)

// NewAllResource is a helper function to simplify the provider implementation.
func NewAllResource() resource.Resource {
	return &allResource{}
}

// allResource is the resource implementation.
type allResource struct {
	owner  string
	client *github.Client
}

// allResourceModel maps the resource schema data.
type allResourceModel struct {
	Repos map[string]allResourceRepoModel `tfsdk:"repos"`
}

// allResourceRepoModel maps repo data.
type allResourceRepoModel struct {
	ID types.Int64 `tfsdk:"id"`
}

func (r *allResource) readRepositories(ctx context.Context, stateRepos *map[string]allResourceRepoModel) error {
	// Get refreshed repositories from GitHub
	tflog.Debug(ctx, "Reading GitHub repositories")

	opt := &github.RepositoryListByOrgOptions{
		Sort:        "full_name",
		ListOptions: github.ListOptions{PerPage: 100},
	}
	var allRepos []*github.Repository
	for {
		repos, gresp, err := r.client.Repositories.ListByOrg(ctx, r.owner, opt)
		if err != nil {
			return err
		}
		allRepos = append(allRepos, repos...)
		if gresp.NextPage == 0 {
			break
		}
		opt.Page = gresp.NextPage
	}
	tflog.Debug(ctx, "GitHub repos are read", map[string]interface{}{"count": len(allRepos)})

	tflog.Debug(ctx, "Setting repos to state")
	for _, repo := range allRepos {
		if _, ok := (*stateRepos)[*repo.Name]; !ok {
			continue
		}
		(*stateRepos)[*repo.Name] = allResourceRepoModel{
			ID: types.Int64Value(*repo.ID),
		}
	}

	tflog.Debug(ctx, "Finished reading GitHub repositories")
	return nil
}

// Configure adds the provider configured client to the resource.
func (r *allResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = config.Client
	r.owner = config.Owner
}

// Metadata returns the resource type name.
func (r *allResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_all"
}

// Schema defines the schema for the resource.
func (r *allResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: descriptions["all_resource_schema"],
		Attributes: map[string]schema.Attribute{
			"repos": schema.MapNestedAttribute{
				Description: descriptions["repos"],
				Required:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Description: descriptions["repo_id"],
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *allResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Get current state
	var state allResourceModel
	diags := req.Plan.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Overwrite items with refreshed state
	err := r.readRepositories(ctx, &state.Repos)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading GitHub Repositories",
			"Could not read GitHub repositories: "+err.Error(),
		)
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *allResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state allResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Overwrite items with refreshed state
	err := r.readRepositories(ctx, &state.Repos)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading GitHub Repositories",
			"Could not read GitHub repositories: "+err.Error(),
		)
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *allResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get current state
	var state allResourceModel
	diags := req.Plan.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Overwrite items with refreshed state
	err := r.readRepositories(ctx, &state.Repos)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading GitHub Repositories",
			"Could not read GitHub repositories: "+err.Error(),
		)
		return
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *allResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

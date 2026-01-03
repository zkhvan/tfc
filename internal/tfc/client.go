package tfc

import (
	"github.com/hashicorp/go-tfe"
)

type Client struct {
	*tfe.Client

	// Re-use a common struct for each service.
	common service

	Organizations *OrganizationsService
	Runs          *RunsService
	Variables     *VariablesService
	Workspaces    *WorkspacesService
}

func NewClient(tfeClient *tfe.Client) *Client {
	c := &Client{
		Client: tfeClient,
	}
	c.common.tfc = c
	c.common.tfe = tfeClient

	c.Organizations = (*OrganizationsService)(&c.common)
	c.Runs = (*RunsService)(&c.common)
	c.Variables = (*VariablesService)(&c.common)
	c.Workspaces = (*WorkspacesService)(&c.common)

	return c
}

type service struct {
	tfc *Client
	tfe *tfe.Client
}

type Pagination struct {
	tfe.Pagination
	ReachedLimit bool
}

type ListOptions struct {
	tfe.ListOptions
	Limit int
}

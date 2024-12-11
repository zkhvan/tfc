package tfc

import (
	"github.com/hashicorp/go-tfe"
)

type Client struct {
	*tfe.Client

	// Re-use a common struct for each service.
	common service

	Workspaces *WorkspacesService
}

func NewClient(tfeClient *tfe.Client) *Client {
	c := &Client{
		Client: tfeClient,
	}
	c.common.tfc = c
	c.common.tfe = tfeClient

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

package tfc

import (
	"github.com/hashicorp/go-tfe"
)

type Client struct {
	*tfe.Client
}

func NewClient(c *tfe.Client) *Client {
	wrapper := &Client{
		Client: c,
	}

	return wrapper
}

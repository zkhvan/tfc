package tfc

import (
	"context"

	"github.com/hashicorp/go-tfe"

	"github.com/zkhvan/tfc/internal/tfc/tfepaging"
)

type VariablesService service

type Variable = tfe.Variable

type VariableListOptions struct {
	ListOptions
}

func (s *VariablesService) List(
	ctx context.Context,
	workspaceID string,
	opts *VariableListOptions,
) ([]*Variable, *Pagination, error) {
	o := tfe.VariableListOptions{}

	f := func(lo tfe.ListOptions) ([]*Variable, *tfe.Pagination, error) {
		o.ListOptions = lo
		result, err := s.tfe.Variables.List(ctx, workspaceID, &o)
		if err != nil {
			return nil, nil, err
		}

		return result.Items, result.Pagination, nil
	}

	current := Pagination{}
	pager := tfepaging.New(f)

	var variables []*Variable
	for i, v := range pager.All() {
		current.Pagination = *pager.Current()

		if opts.Limit == 0 {
			opts.Limit = 20
		}

		if opts.Limit <= len(variables) {
			if i < current.TotalCount {
				current.ReachedLimit = true
			}
			break
		}

		variables = append(variables, v)
	}

	if err := pager.Err(); err != nil {
		return nil, nil, err
	}

	return variables, &current, nil
}

func (s *VariablesService) Read(
	ctx context.Context,
	workspaceID string,
	variableID string,
) (*Variable, error) {
	return s.tfe.Variables.Read(ctx, workspaceID, variableID)
}

func (s *VariablesService) Create(
	ctx context.Context,
	workspaceID string,
	options tfe.VariableCreateOptions,
) (*Variable, error) {
	return s.tfe.Variables.Create(ctx, workspaceID, options)
}

func (s *VariablesService) Update(
	ctx context.Context,
	workspaceID string,
	variableID string,
	options tfe.VariableUpdateOptions,
) (*Variable, error) {
	return s.tfe.Variables.Update(ctx, workspaceID, variableID, options)
}

func (s *VariablesService) Delete(
	ctx context.Context,
	workspaceID string,
	variableID string,
) error {
	return s.tfe.Variables.Delete(ctx, workspaceID, variableID)
}

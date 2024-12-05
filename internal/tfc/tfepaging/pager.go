package tfepaging

import (
	"iter"

	"github.com/hashicorp/go-tfe"
)

type Pager[V any] struct {
	f       func(tfe.ListOptions) ([]V, *tfe.Pagination, error)
	opts    *tfe.ListOptions
	index   int
	current *tfe.Pagination
	err     error
}

func New[V any](f func(tfe.ListOptions) ([]V, *tfe.Pagination, error)) *Pager[V] {
	return &Pager[V]{
		f:    f,
		opts: &tfe.ListOptions{},
	}
}

func (p *Pager[V]) SetPageSize(size int) *Pager[V] {
	p.opts.PageSize = size
	return p
}

func (p *Pager[V]) Current() *tfe.Pagination {
	return p.current
}

func (p *Pager[V]) Err() error {
	return p.err
}

func (p *Pager[V]) All() iter.Seq2[int, V] {
	return func(yield func(int, V) bool) {
		for {
			var out []V
			out, p.current, p.err = p.f(*p.opts)
			if p.err != nil {
				return
			}

			for _, v := range out {
				if !yield(p.index, v) {
					return
				}
				p.index++
			}

			if p.current.NextPage == 0 {
				return
			}

			p.opts.PageNumber = p.current.NextPage
		}
	}
}

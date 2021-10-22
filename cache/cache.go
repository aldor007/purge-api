package cache

import "context"

type Cache interface {
	Purge(ctx context.Context, url string) error
}

type Purger struct {
	caches []Cache
}

func NewPurger() Purger {
	return Purger{caches: []Cache{}}
}

func (p *Purger) AddCache(c Cache)  {
	p.caches = append(p.caches, c)
}

func (p *Purger) Purge(ctx context.Context, url string) error  {
	var err error
	for _, c := range p.caches {
		errP  := c.Purge(ctx, url)
		if errP != nil {
			err = errP
		}

	}

	return err
}
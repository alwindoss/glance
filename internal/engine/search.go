package engine

import "github.com/alwindoss/glance/internal/search"

func Search(dir, query string, parallel int) error {
	search.Search(dir, query, parallel)
	return nil
}

package engine

type Finder interface {
	Find(query string) error
}

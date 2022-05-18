package state

type List struct {
	ls []string
}

func NewList(name string) *List {
	ls := &List{ls: make([]string, 0)}
	return ls
}

// Push is a list abstraction layered over the same table. Pushing to the list
// is the same as enqueuing w.r.t. the DB.
func (ls *List) Push(s string) {
	ls.ls = append([]string{s}, ls.ls...)
}

func (ls *List) AsList() []string {
	return ls.ls
}

func (ls *List) Remove(s string) {
	n := make([]string, 0)
	for _, v := range ls.ls {
		if v != s {
			n = append(n, v)
		}
	}
	ls.ls = n
}

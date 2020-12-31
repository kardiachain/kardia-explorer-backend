package types

const (
	defaultLimit = 50
	MaximumLimit = 100
)

type Pagination struct {
	Skip  int
	Limit int
}

func (f *Pagination) Sanitize() {
	if f.Skip < 0 {
		f.Skip = 0
	}
	if f.Limit <= 0 {
		f.Limit = defaultLimit
	} else if f.Limit > MaximumLimit {
		f.Limit = MaximumLimit
	}
}

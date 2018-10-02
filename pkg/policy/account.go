package policy

// Account is an aws account. It should probably be in pkg/aws, but then we end up with a cycle.
type Account struct {
	Name  string
	ID    int64
	Role  string
	Owner string
}

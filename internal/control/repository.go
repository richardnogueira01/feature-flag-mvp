package control

import "context"

// Repository is the persistence port required by the control plane.
// Implementations must keep mutation and outbox insertion in one transaction.
type Repository interface {
	Apply(context.Context, string, bool) (Flag, error)
	Delete(context.Context, string) error
	Get(context.Context, string) (Flag, error)
	List(context.Context) ([]Flag, error)
}

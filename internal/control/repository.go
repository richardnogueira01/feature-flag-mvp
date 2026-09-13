package control

import (
	"context"
	"encoding/json"
)

type Repository interface {
	Apply(context.Context, string, bool) (Flag, error)
	ApplyValue(context.Context, string, json.RawMessage) (Flag, error)
	Delete(context.Context, string) error
	Get(context.Context, string) (Flag, error)
	List(context.Context) ([]Flag, error)
}

package cmd

import (
	"context"

	"github.com/mircearoata/wwise-cli/lib/wwise/client"
)

type key int

var wwiseClient key

func NewContextWithClient(ctx context.Context, c *client.WwiseClient) context.Context {
	return context.WithValue(ctx, wwiseClient, c)
}

// FromContext returns the User value stored in ctx, if any.
func ClientFromContext(ctx context.Context) (*client.WwiseClient, bool) {
	u, ok := ctx.Value(wwiseClient).(*client.WwiseClient)
	return u, ok
}

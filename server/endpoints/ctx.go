package endpoints

import (
	"context"
	"fmt"
	"strafe/internal"
)

func getAppContext(ctx context.Context) (internal.AppCtx, error) {
	app, ok := ctx.Value(internal.APP_CONTEXT_KEY).(internal.AppCtx)
	if !ok {
		return internal.AppCtx{}, fmt.Errorf("app context not found in context")
	}
	return app, nil
}

package command

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var EmitEvent = func(ctx context.Context, eventName string, data ...interface{}) {
	runtime.EventsEmit(ctx, eventName, data...)
}
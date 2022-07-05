package utils

import (
	"context"
	"google.golang.org/grpc/metadata"
	"strings"
)

const TraceparentKey = "traceparent"

func GetTraceLogEntryFromContext(ctx context.Context) map[string]interface{} {
	ctx.Value("trace_id")
	return map[string]interface{}{
		"trace_id": ctx.Value("trace_id"),
		"span_id":  ctx.Value("span_id"),
	}
}

type SpanContext struct {
	TraceID      string
	SpanID       string
	TraceOptions string
}

//从传入上下文获取trace_id
func GetTraceInfoFromCtx(ctx context.Context) SpanContext {
	mCtx, ok := metadata.FromIncomingContext(ctx)
	if ok {
		traceparents := mCtx.Get(TraceparentKey)
		if len(traceparents) > 0 {
			traceStr := traceparents[0]
			traceSlice := strings.Split(traceStr, "-")
			if len(traceSlice) == 4 {
				return SpanContext{
					TraceID:      traceSlice[1],
					SpanID:       traceSlice[2],
					TraceOptions: traceSlice[3],
				}
			}
		}
	}
	return SpanContext{}
}

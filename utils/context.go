package utils

import (
	"context"
	"google.golang.org/grpc/metadata"
	"strings"
)

const TraceparentKey = "traceparent"

func GetCommonMetaFromCtx(ctx context.Context) map[string]interface{} {
	meta := make(map[string]interface{})
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		//dapr trace处理
		traceparents := md.Get(TraceparentKey)
		if len(traceparents) > 0 {
			traceStr := traceparents[0]
			traceSlice := strings.Split(traceStr, "-")
			if len(traceSlice) == 4 {
				meta["trace_id"] = traceSlice[1]
				meta["span_id"] = traceSlice[2]
			}
		}
		//xxl log_id处理
		logIds := md.Get("xxl_log_id")
		if len(logIds) > 0 {
			meta["log_id"] = logIds[0]
		}
	}
	return meta
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

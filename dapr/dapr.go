package dapr

import (
	bytes2 "bytes"
	"compress/zlib"
	"context"
	"encoding/json"
	"github.com/dapr/go-sdk/client"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang/protobuf/proto"
	"github.com/powershopy/library/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

var cli client.Client
var status = false
var wg = sync.WaitGroup{}

const (
	daprPortDefault    = "50001"
	daprPortEnvVarName = "DAPR_GRPC_PORT" /* #nosec */
)

func init() {
	go func() {
		var err error
		wg.Add(1)
		defer wg.Done()
		retry := 0
	Init:
		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second) //等待dapr初始化
		port := os.Getenv(daprPortEnvVarName)
		if port == "" {
			port = daprPortDefault
		}
		address := net.JoinHostPort("127.0.0.1", port)
		var kacp = keepalive.ClientParameters{
			Time:                5 * time.Second, // send pings every 10 seconds if there is no activity
			Timeout:             time.Second,     // wait 1 second for ping ack before considering the connection dead
			PermitWithoutStream: true,            // send pings even without active streams
		}
		conn, err := grpc.DialContext(
			ctx,
			address,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithKeepaliveParams(kacp),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*100), grpc.MaxCallSendMsgSize(1024*1024*100)),
		)
		if err != nil && retry >= 3 {
			log.Fatalln(err)
		} else if err != nil {
			retry++
			goto Init
		}
		cli = client.NewClientWithConnection(conn)
		if cli == nil {
			log.Fatalln(err)
		}
		status = true
	}()
}

func Status() bool {
	return status
}

func PubSubMethod(ctx context.Context, appName, PubSubName, topicName string, data interface{}) error {
	wg.Wait()
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	traceId := ghttp.RequestFromCtx(ctx).Header.Get("traceparent")
	logging.WithFields(map[string]interface{}{
		"app_id":      appName,
		"pubsub_name": PubSubName,
		"topic_name":  topicName,
	}).Info(ctx, "start pubsub")
	ctx = context.Background()
	if traceId != "" {
		ctx = cli.WithTraceID(ctx, traceId) //链路追踪
	}
	return cli.PublishEvent(ctx, PubSubName, topicName, bytes)
}

func InvokeMethod(ctx context.Context, appName, method string, data interface{}) ([]byte, error) {
	wg.Wait()
	bytes, err := json.Marshal(data)
	if err != nil {
		return []byte{}, err
	}
	content := &client.DataContent{
		ContentType: "application/json",
		Data:        bytes,
	}
	traceId := ghttp.RequestFromCtx(ctx).Header.Get("traceparent")
	logging.WithFields(map[string]interface{}{
		"app_id": appName,
		"method": method,
	}).Info(ctx, "start invoke")
	ctx = context.Background()
	if traceId != "" {
		ctx = cli.WithTraceID(ctx, traceId) //链路追踪
	}
	res, err := cli.InvokeMethodWithContent(ctx, appName, method, "post", content)
	if len(res) > 0 && res[0] != '{' { //解压缩数据
		var out bytes2.Buffer
		r, e := zlib.NewReader(bytes2.NewReader(res))
		if e != nil {
			return res, e
		}
		_, e = io.Copy(&out, r)
		if e != nil {
			return res, e
		}
		res = out.Bytes()
	}
	//增加调用错误日志
	if err != nil {
		logging.WithFields(map[string]interface{}{
			"app_id": appName,
			"method": method,
			"err":    err,
		}).Error(ctx, "invoke err")
	}
	return res, err
}

func InvokeMethodWithProto(ctx context.Context, appName, method string, request, response proto.Message) error {
	wg.Wait()
	data, err := proto.Marshal(request)
	if err != nil {
		return err
	}
	content := &client.DataContent{
		Data:        data,
		ContentType: "protobuf",
	}
	traceId := ghttp.RequestFromCtx(ctx).Header.Get("traceparent")
	logging.WithFields(map[string]interface{}{
		"app_id": appName,
		"method": method,
	}).Info(ctx, "start invoke")
	ctx = context.Background()
	if traceId != "" {
		ctx = cli.WithTraceID(ctx, traceId) //链路追踪
	}
	out, err := cli.InvokeMethodWithContent(ctx, appName, method, "post", content)
	//增加调用错误日志
	if err != nil {
		logging.WithFields(map[string]interface{}{
			"app_id": appName,
			"method": method,
			"err":    err,
		}).Error(ctx, "invoke err")
		return err
	}
	err = proto.Unmarshal(out, response)
	return err
}

func Client() client.Client {
	return cli
}
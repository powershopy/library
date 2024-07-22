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
	"strings"
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
		logging.Info(context.Background(), "start init daprd")
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
			Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
			Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
			PermitWithoutStream: true,             // send pings even without active streams
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
			logging.WithFields(map[string]interface{}{
				"retry": retry,
				"err":   err,
			}).Info(context.Background(), "dapr dial err")
			log.Fatalln(err)
		} else if err != nil {
			logging.WithFields(map[string]interface{}{
				"retry": retry,
				"err":   err,
			}).Info(context.Background(), "dapr dial err")
			retry++
			goto Init
		}
		logging.Info(context.Background(), "dapr dial success")
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
	r := ghttp.RequestFromCtx(ctx)
	if r != nil {
		traceId := ghttp.RequestFromCtx(ctx).Header.Get("traceparent")
		logging.WithFields(map[string]interface{}{
			"app_id":      appName,
			"pubsub_name": PubSubName,
			"topic_name":  topicName,
		}).Info(ctx, "start pubsub")
		if traceId != "" {
			ctx = cli.WithTraceID(ctx, traceId) //链路追踪
		}
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
	r := ghttp.RequestFromCtx(ctx)
	if r != nil {
		traceId := ghttp.RequestFromCtx(ctx).Header.Get("traceparent")
		logging.WithFields(map[string]interface{}{
			"app_id": appName,
			"method": method,
		}).Info(ctx, "start invoke")
		if traceId != "" {
			ctx = cli.WithTraceID(ctx, traceId) //链路追踪
		}
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

func InvokeMethodWithProto(ctx context.Context, appName, method string, request, response proto.Message, retry ...int64) error {
	wg.Wait()
	data, err := proto.Marshal(request)
	if err != nil {
		return err
	}
	content := &client.DataContent{
		Data:        data,
		ContentType: "protobuf",
	}
	r := ghttp.RequestFromCtx(ctx)
	if r != nil {
		traceId := ghttp.RequestFromCtx(ctx).Header.Get("traceparent")
		logging.WithFields(map[string]interface{}{
			"app_id": appName,
			"method": method,
		}).Info(ctx, "start invoke")
		if traceId != "" {
			ctx = cli.WithTraceID(ctx, traceId) //链路追踪
		}
	}
Invoke:
	out, err := cli.InvokeMethodWithContent(ctx, appName, method, "post", content)
	//增加调用错误日志
	if err != nil {
		if len(retry) > 0 && retry[0] > 0 && strings.Contains(err.Error(), "error reading from server") {
			//因为服务器压力扩容导致的访问不可用，增加可选重试机制
			retry[0]--
			time.Sleep(time.Millisecond * 500)
			goto Invoke
		}
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

type NormalDaprRes struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func PublishEvent(ctx context.Context, pubsubName, topicName string, data interface{}) error {
	wg.Wait()
	r := ghttp.RequestFromCtx(ctx)
	if r != nil {
		traceId := ghttp.RequestFromCtx(ctx).Header.Get("traceparent")
		if traceId != "" {
			ctx = cli.WithTraceID(ctx, traceId) //链路追踪
		}
	}
	err := cli.PublishEvent(ctx, pubsubName, topicName, data)
	//增加调用错误日志
	if err != nil {
		logging.WithFields(map[string]interface{}{
			"pubsubName": pubsubName,
			"topicName":  topicName,
			"err":        err,
		}).Error(ctx, "publish err")
		return err
	}
	return err
}

func InvokeOutputBinding(ctx context.Context, in *client.InvokeBindingRequest) error {
	r := ghttp.RequestFromCtx(ctx)
	if r != nil {
		traceId := ghttp.RequestFromCtx(ctx).Header.Get("traceparent")
		if traceId != "" {
			ctx = cli.WithTraceID(ctx, traceId) //链路追踪
		}
	}
	err := cli.InvokeOutputBinding(ctx, in)
	return err
}

func Client() client.Client {
	return cli
}

// 关闭边车
func Shutdown(ctx context.Context) error {
	var err error
	for i := 0; i < 20; i++ {
		err = cli.Shutdown(ctx)
		logging.WithFields(map[string]interface{}{
			"err":         err,
			"retry_times": i,
		}).Info(context.Background(), "shutdown sidecar result")
		if err != nil {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	return err
}

// s3文件上传
func Upload(ctx context.Context, componentName string, fileKey string, filePath string) error {
	wg.Wait()
	err := cli.InvokeOutputBinding(ctx, &client.InvokeBindingRequest{
		Name:      componentName,
		Operation: "create",
		Metadata: map[string]string{
			"key":      fileKey,
			"filePath": filePath,
		},
	})
	return err
}

func Download(ctx context.Context, componentName string, fileKey string) ([]byte, error) {
	wg.Wait()
	out, err := cli.InvokeBinding(ctx, &client.InvokeBindingRequest{
		Name:      componentName,
		Operation: "get",
		Metadata: map[string]string{
			"key": fileKey,
		},
	})
	if err != nil {
		return nil, err
	}
	return out.Data, err
}

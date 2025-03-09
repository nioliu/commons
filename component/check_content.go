package component

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/nioliu/commons/log"
	"github.com/nioliu/protocols/httpproto"
	"go.uber.org/zap"
	"io"
	"net/http"
)

const dataServiceUrl = "http://data-service:8080"

func CheckContent(ctx context.Context, req *httpproto.CheckContentReq) (*httpproto.CheckContentRsp, error) {

	baseURL := dataServiceUrl + "/v1/data/content/check"
	reqBytes, err := json.Marshal(req)
	if err != nil {
		log.ErrorWithCtxFields(ctx, "marshal request failed", zap.Error(err))
		return nil, err
	}

	// 创建请求
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, bytes.NewReader(reqBytes))
	if err != nil {
		log.ErrorWithCtxFields(ctx, "create request failed", zap.Error(err))
		return nil, fmt.Errorf("create request failed: " + err.Error())
	}

	// 设置请求头
	request.Header.Set("Content-Type", "application/json")

	// 发送请求
	rsp, err := http.DefaultClient.Do(request)
	if err != nil {
		log.ErrorWithCtxFields(ctx, "do request failed", zap.Error(err))
		return nil, fmt.Errorf("do request failed: " + err.Error())
	}
	defer rsp.Body.Close()

	// 检查响应状态码
	if rsp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(rsp.Body)
		log.ErrorWithCtxFields(ctx, "request failed with non-200 status code",
			zap.Int("status_code", rsp.StatusCode),
			zap.String("response_body", string(body)))
		return nil, fmt.Errorf(fmt.Sprintf("request failed with status code: %d", rsp.StatusCode))
	}

	// 读取响应
	h := new(httpproto.CheckContentRsp)
	if err = json.NewDecoder(rsp.Body).Decode(h); err != nil {
		log.ErrorWithCtxFields(ctx, "decode response body failed", zap.Error(err))
		return nil, fmt.Errorf("decode response body failed: " + err.Error())
	}

	return h, nil
}

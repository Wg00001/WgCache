package wgCacheHttp

import (
	protocol "WgCache/wgCacheProto"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io"
	"net/http"
	"net/url"
)

type httpGetter struct {
	baseURL string
}

func (g *httpGetter) Get(in *protocol.Request, out *protocol.Response) error {
	//构造url，url.QueryEscape()用于对字符串进行转义，以便可以将其安全地放置在URL查询中。
	u := fmt.Sprintf("%v/%v/%v", g.baseURL, url.QueryEscape(in.GetGroup()), url.QueryEscape(in.GetKey()))
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("ERR: reading response body: %v", err)
	}
	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}

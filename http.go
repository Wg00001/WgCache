package wgCache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultPath = "/wgCache/"

type HTTPPool struct {
	selfPath string //自己的地址
	basePath string //节点间通讯地址的前缀
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		selfPath: self,
		basePath: defaultPath,
	}
}

func (p *HTTPPool) Log(formate string, v ...interface{}) {
	log.Printf("[Server %s]: %s", p.selfPath, fmt.Sprintf(formate, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("ERR:http.ServeHTTP:HTTPPool serving unexpected path:" + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	//约定访问路径格式为：/<basepath>/<groupname>/<key>
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

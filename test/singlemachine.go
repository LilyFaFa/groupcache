package main

import (
	"errors"
	//"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	groupcache "github.com/golang/groupcache"
)

const (
	DefaulPort = ":9000"
)

/*
	创建一个简单的数据源获取方式，直接拿自己struct内的内容，如果cache中不存在就会查询kv
	在实际使用的时候可以查询其他的数据源
*/
type getFunc struct {
	kv map[string]string
}

func (g *getFunc) Get(ctx groupcache.Context, key string, dest groupcache.Sink) error {
	value, ok := g.kv[key]
	if ok {
		//储存的数据是string，所以使用SetString函数
		dest.SetString(value)
		/*
			The cache will ask once if it has the key-value
		*/
		log.Print("Get value from database.")
		return nil
	}
	return errors.New("Error happened finding key")
}

func main() {
	/*
		支持自己设置监听端口，默认值为9000
	*/
	var port string
	if len(os.Args) < 2 {
		port = DefaulPort
	} else {
		port = ":" + os.Args[1]
	}
	/*
		新建一个分组，参数顺序(Group的名称，cache大小，Getter)
	*/
	value := make(map[string]string)
	value["key"] = "value"
	lily := groupcache.NewGroup("lily", 1024*1024*1024*16, &getFunc{value})

	/*
		启动一个http服务
	*/
	http.HandleFunc("/singlemachine/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.URL.Path[len("/singlemachine/"):], "/", 1)
		if len(parts) != 1 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		var datas []byte
		/*
			获取数据
		*/
		lily.Get(nil, parts[0], groupcache.AllocatingByteSliceSink(&datas))
		w.Write(datas)
		log.Print(string(datas))
	})

	http.ListenAndServe(port, nil)
}

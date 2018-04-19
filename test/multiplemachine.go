package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	groupcache "github.com/golang/groupcache"
)

const (
	DefaulPort = ":9000"
)

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
	peers := groupcache.NewHTTPPool("http://localhost" + port)
	peers.Set("http://127.0.0.1:9001", "http://127.0.0.1:9002")
	/*
		rpcPeers := []string{
			"http://127.0.0.1:8000", "http://127.0.0.1:8001", "http://127.0.0.1:8002",
		}
	*/
	/*
		新建一个分组，参数顺序(Group的名称，cache大小，Getter)
	*/
	value := make(map[string]string)
	value["key"] = "value"
	lily := groupcache.NewGroup("lily", 1024*1024*1024*16, &getFunc{value})

	go http.ListenAndServe(":9000", peers)
	go http.ListenAndServe(":9001", peers)
	go http.ListenAndServe(":9002", peers)
	go server(lily, ":9001")
	go server(lily, ":9002")
	server(lily, port)
	fmt.Println("hello,here")
	/*
		启动一个http服务
	*/

}
func server(g *groupcache.Group, port string) {

	fmt.Println("hello,haha", port)
	http.HandleFunc("/singlemachine/"+port+"/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.SplitN(r.URL.Path[len("/singlemachine/"+port+"/"):], "/", 1)
		if len(parts) != 1 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		fmt.Println(parts[0])
		var datas []byte
		/*
			获取数据
		*/
		err := g.Get(nil, parts[0], groupcache.AllocatingByteSliceSink(&datas))
		if err != nil {
			fmt.Println("&&&&&&&&")
			w.Write([]byte(err.Error()))
		} else {
			fmt.Println("*******")
			w.Write(datas)
		}

		log.Print(string(datas))
	})

	http.ListenAndServe(port, nil)
}

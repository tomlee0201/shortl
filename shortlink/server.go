package shortlink

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/spf13/viper"
	"github.com/tomlee0201/shortl/shortlink/db"
)

var domain string
var port string

func recordAccess(key string, ua string, ip string, status int) {
	go db.InsertAccessRecord(key, ua, ip, status)
}
func handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if p := recover(); p != nil {
			fmt.Printf("panic recover! p: %v", p)
			debug.PrintStack()
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Server internal error")
		}
	}()

	key := r.URL.Path[1:]
	if key == "" {
		createHandler(w, r)
		return
	}
	entry, error := db.Query(key)
	if error == nil {
		if entry.Duration != 0 {
			ct := time.Now().Unix()
			if ct > entry.CreateAt+entry.Duration {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "时间过期")
				recordAccess(entry.Key, r.UserAgent(), r.RemoteAddr, 2)
				return
			}
		}

		if entry.Password != "" {
			password := r.URL.Query().Get("pwd")
			if password == "" {
				fmt.Fprintf(w, "<!DOCTYPE html>"+
					"<html> "+
					"<head><title>请输入密码</title><meta http-equiv='Content-Type' content='text/html; charset=utf-8'></head> "+
					"<body><h1>请输入访问密码</h1>密码：<input type='text' id='pwd' /></br></br><button type='button' onclick='myFunction()'>Go</button><script> function myFunction(){var elementPwd = document.getElementById('pwd'); var url = self.location.href;if (elementPwd.value != '') {window.location.href = url + '?pwd=' + elementPwd.value;}}</script></body> "+
					"</html>")
				return
			}
			if password != entry.Password {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "密码错误")
				recordAccess(entry.Key, r.UserAgent(), r.RemoteAddr, 3)
				return
			}
			r.URL.Query().Del("pwd")
		}

		http.Redirect(w, r, entry.Url, http.StatusFound)
		recordAccess(entry.Key, r.UserAgent(), r.RemoteAddr, 0)
		return
	} else {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Invalide url")
		recordAccess(entry.Key, r.UserAgent(), r.RemoteAddr, 1)
	}
}

type CreateResult struct {
	Orignal string `json:"orignal"`
	Key     string `json:"key"`
	Domain  string `json:"domain""`
	Port    string `json:"port"`
}

func createShortLink(url string, duration int64, password string) CreateResult {
	key := db.InsertWithRetGenKey(url, duration, password)
	result := CreateResult{url, key, domain, port}
	return result
}

func createApiHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		url := r.PostFormValue("url")
		sduration := r.PostFormValue("duration")
		password := r.PostFormValue("password")

		duration, error := strconv.ParseInt(sduration, 10, 64)
		if error != nil {
			duration = 0
		}

		result := createShortLink(url, duration, password)
		js, err := json.Marshal(result)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalide method, only POST support")
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<!DOCTYPE html>"+
		"<html>"+
		"<head> <title>短链生成器</title> <meta http-equiv='Content-Type' content='text/html; charset=utf-8'> <script src='http://ajax.aspnetcdn.com/ajax/jQuery/jquery-1.8.0.js'> </script> </head> "+
		"<body> <h1>短链生成器</h1> <p id='demo'>输入Url，点击确认，生成短链</p> URL：<input type='text' id='url' /> 有效期(秒)：<input type='text' size = '6' id='duration' /> 密码：<input type='text' size = '10' id='pwd' /> </br></br> <button type='button' onclick='myFunction()'>确认</button></br></br> 结果: <a href='' id='resulta'></a>"+
		"</br></br><a href='https://github.com/tomlee0201/shortl'>源码</a>"+
		"<script> function myFunction(){ var resulta = document.getElementById('resulta'); var elementUrl = document.getElementById('url'); var elementDuration = document.getElementById('duration'); var elementPwd = document.getElementById('pwd'); resulta.innerHTML = ''; resulta.setAttribute('href', '');   $.post('/api/create',{url:elementUrl.value,duration:elementDuration.value,password:elementPwd.value}, function(data,status){if(status = 'success') { var shortUrl = 'http://' + data.domain; if(data.port != '80') { shortUrl = shortUrl + ':' + data.port; } shortUrl = shortUrl + '/' + data.key; resulta.innerHTML = shortUrl;resulta.setAttribute('href', shortUrl); if(elementDuration.value != ''){alert('请在 ' + elementDuration.value + 's 内访问')}} else {alert('错误：' + data);}}); }</script>"+
		"</body> </html> ")
}

func StartServer() {
	db.Init()

	domain = viper.GetString("services.shortlink.domain")
	port = viper.GetString("services.shortlink.port")
	http.HandleFunc("/api/create", createApiHandler)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+port, nil)
}

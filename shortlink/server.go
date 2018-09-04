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
	"io/ioutil"
	"strings"
)

var domain string
var port string

var createHtml string
var authHtml string
var pwdErrorHtml string
var expiredHtml string
var notExistHtml string
var serverErrorHtml string

func recordAccess(key string, ua string, ip string, status int) {
	go db.InsertAccessRecord(key, ua, ip, status)
}
func handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if p := recover(); p != nil {
			fmt.Printf("panic recover! p: %v", p)
			debug.PrintStack()
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, serverErrorHtml)
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
				fmt.Fprint(w, expiredHtml)
				recordAccess(entry.Key, r.UserAgent(), r.RemoteAddr, 2)
				return
			}
		}

		if entry.Password != "" {
			password := r.URL.Query().Get("pwd")
			if password == "" {
				fmt.Fprint(w, authHtml)
				return
			}
			if password != entry.Password {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprint(w, pwdErrorHtml)
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
		fmt.Fprint(w, notExistHtml)
		recordAccess(key, r.UserAgent(), r.RemoteAddr, 1)
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
		if !strings.Contains(url, ":") {
			url = "http://" + url
		}

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
		fmt.Fprint(w, "Invalide method, only POST support")
	}
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, createHtml)
}

func StartServer() {
	db.Init()
	b, err := ioutil.ReadFile("./resource/create.html")
	if err != nil {
		fmt.Print(err)
	}

	createHtml = string(b)

	c, err := ioutil.ReadFile("./resource/auth.html")
	if err != nil {
		fmt.Print(err)
	}

	authHtml = string(c)

	d, err := ioutil.ReadFile("./resource/error.html")
	if err != nil {
		fmt.Print(err)
	}

	pwdErrorHtml = string(d)
	expiredHtml = strings.Replace(pwdErrorHtml, "密码错误", "过期了。。。", 1)
	notExistHtml = strings.Replace(pwdErrorHtml, "密码错误", "网址不存在", 1)
	serverErrorHtml = strings.Replace(pwdErrorHtml, "密码错误", "服务器错误", 1)


	domain = viper.GetString("services.shortlink.domain")
	port = viper.GetString("services.shortlink.port")
	http.HandleFunc("/api/create", createApiHandler)
	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+port, nil)
}

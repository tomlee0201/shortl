package db

import (
	"database/sql"
	"fmt"
	"regexp"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"github.com/tomlee0201/shortl/util"
)

var lock sync.Mutex
var cache *util.LRUCache
var db *sql.DB

type Error struct {
	S string
}

func (e *Error) Error() string {
	return e.S
}

type Entry struct {
	Key      string
	Url      string
	Duration int64
	Password string
	CreateAt int64
}

type AccessRecord struct {
	Key       string
	UserAgent string
	FromIp    string
	CreateAt  int64
	Status    int //0 normal; 1 no entry; 2 time expired; 3 pwd error; 4 deleted
}

func Init() {
	cacheCap := viper.GetInt("services.shortlink.lru_cache_size")
	cache = util.NewLRUCache(cacheCap)

	dbUrl := viper.GetString("services.shortlink.db")

	var err error
	db, err = sql.Open("mysql", dbUrl)
	if err != nil {
		fmt.Println("open db failure ", err)
	}

}

func Query(key string) (*Entry, error) {
	lock.Lock()
	defer lock.Unlock()

	exp1 := regexp.MustCompile("[a-zA-Z0-9\\-_]+")

	marched := exp1.FindAllString(key, -1)
	if marched == nil || marched[0] != key {
		return nil, &Error{S: "Invalide key " + key}
	}
	v, succ, _ := cache.Get(key)
	if succ {
		value, ok := v.(Entry)
		if ok {
			return &value, nil
		}
	}

	if db != nil {
		rows, err := db.Query(`SELECT _value, _duration, _password, _dt FROM t_entry where _key = '` + key + `' limit 1`)
		if checkErr(err) {
			return nil, &Error{S: "DB error "}
		}

		if rows.Next() {
			var value string
			var duration int64
			var password string
			var dt int64
			rows.Columns()
			err = rows.Scan(&value, &duration, &password, &dt)
			if err == nil {
				entry := Entry{Key: key, Url: value, Duration: duration, Password: password, CreateAt: dt}
				cache.Set(key, entry)
				return &entry, nil
			}
		}
	}

	return nil, &Error{S: "Invalide url " + key}
}

func InsertWithRetGenKey(value string, duration int64, password string) string {
	lock.Lock()
	defer lock.Unlock()

	dt := time.Now().Unix()

	key, err := util.Generate()
	if err == nil {
		cache.Set(key, Entry{Key: key, Url: value, Duration: duration, Password: password, CreateAt: dt})
	}

	go func() {
		if db != nil {
			stmt, err := db.Prepare(`INSERT t_entry (_key, _value, _duration, _password, _dt) values (?,?,?,?,?)`)
			if checkErr(err) {
				return
			}
			res, err := stmt.Exec(key, value, duration, password, dt)
			if checkErr(err) {
				return
			}
			id, err := res.LastInsertId()
			if checkErr(err) {
				return
			}
			fmt.Println("insert to entry table with id ", id)
		}
	}()

	return key
}

func InsertAccessRecord(key string, ua string, ip string, status int) {
	dt := time.Now().Unix()

	if db != nil {
		stmt, err := db.Prepare(`INSERT t_access_record (_key, _ua, _ip, _status, _dt) values (?,?,?,?,?)`)
		if checkErr(err) {
			return
		}
		res, err := stmt.Exec(key, ua, ip, status, dt)
		if checkErr(err) {
			return
		}
		id, err := res.LastInsertId()
		if checkErr(err) {
			return
		}
		fmt.Println("insert to access record table with id ", id)
	}
}

func checkErr(err error) bool {
	if err != nil {
		fmt.Println(err)
		return true
	}
	return false
}

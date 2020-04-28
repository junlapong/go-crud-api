package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	//_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

const (
	//mysql db setting
	driver = "sqlite3"
	//user     = "root"
	//password = "r00tp@55"
	//host     = "127.0.0.1"
	//port     = "3306"
	database = "db/crud.sqlite"

	//server setting
	serverPort = "8000"

	maxConnections = 256
)

var (
	db *sql.DB
)

var (
	listenAddr = flag.String("listenAddr", ":"+serverPort, "Address to listen to")
	//connString = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, database)
	connString = fmt.Sprintf("%s", database)
)

func requestHandler(w http.ResponseWriter, req *http.Request) {
	var msg []byte
	var data interface{}
	w.Header().Add("Content-Type", "application/json")

	method := req.Method
	u, _ := url.ParseRequestURI(req.RequestURI)
	request := strings.Split(strings.Trim(u.Path, "/"), "/")

	// load input from request body
	var input map[string]interface{}
	decoder := json.NewDecoder(req.Body)
	decoder.Decode(&input)

	// retrieve the table and key from the path
	table := regexp.MustCompile("[^a-z0-9_]+").ReplaceAllString(request[0], "")
	key := 0
	if len(request) > 1 {
		key, _ = strconv.Atoi(request[1])
	}

	// escape the columns from the input object
	var args []interface{}
	if key > 0 {
		args = make([]interface{}, 0, len(input)+1)
	} else {
		args = make([]interface{}, 0, len(input))
	}

	columns := make([]string, 0, len(input))
	postColumns := make([]string, 0, len(input))
	postValues := make([]string, 0, len(input))

	for column, arg := range input {
		name := regexp.MustCompile("[^a-z0-9_]+").ReplaceAllString(column, "")
		args = append(args, arg)
		columns = append(columns, fmt.Sprintf("`%s`=?", name))
		postColumns = append(postColumns, name)
		postValues = append(postValues, "?")
	}

	set := strings.Join(columns, ", ")
	postC := strings.Join(postColumns, ", ")
	postV := strings.Join(postValues, ", ")

	if key > 0 {
		args = append(args, key)
	}

	// create SQL based on HTTP method
	query := ""
	switch method {
	case "GET":
		if key > 0 {
			query = fmt.Sprintf("select * from `%s` where `id`=?", table)
		} else {
			query = fmt.Sprintf("select * from `%s`", table)
		}
		break
	case "PUT":
		query = fmt.Sprintf("update `%s` set %s where `id`=?", table, set)
		break
	case "POST":
		query = fmt.Sprintf("insert into `%s`(%s) values(%s)", table, postC, postV)
		break
	case "DELETE":
		query = fmt.Sprintf("delete from `%s` where `id`=?", table)
		break
	}

	log.Println("SQL: " + query)

	if method == "GET" {
		rows, err := db.Query(query, args...)
		checkErr(err)

		cols, err := rows.Columns()
		checkErr(err)

		if key == 0 {
			w.Write([]byte("{"))
			msg, _ = json.Marshal(table)
			w.Write(msg)
			w.Write([]byte(":{\"columns\":"))
			msg, _ = json.Marshal(cols)
			w.Write(msg)
			w.Write([]byte(",\"records\":["))
		}

		values := make([]interface{}, len(cols))
		record := make(map[string]interface{})
		for i, col := range cols {
			var value *string
			values[i] = &value
			record[col] = &value
		}

		for i := 0; rows.Next(); i++ {
			err := rows.Scan(values...)
			checkErr(err)

			if key == 0 {
				if i > 0 {
					w.Write([]byte(","))
				}
				msg, _ = json.Marshal(values)
				w.Write(msg)
			} else {
				msg, _ = json.Marshal(record)
				w.Write(msg)
			}
		}
		if key == 0 {
			w.Write([]byte("]}}"))
		}
	} else if method == "POST" {
		stmt, err := db.Prepare(query)
		checkErr(err)

		res, err := stmt.Exec(args...)
		checkErr(err)

		id, err := res.LastInsertId()
		checkErr(err)

		msg, _ = json.Marshal(id)
		w.Write(msg)

	} else {
		result, err := db.Exec(query, args...)
		checkErr(err)

		data, _ = result.RowsAffected()
		msg, _ = json.Marshal(data)
		w.Write(msg)
	}
}

func checkErr(err error) {
	if err != nil {
		log.Print(err)
		return
	}
}

func main() {
	var err error
	db, err = sql.Open(driver, connString)
	if err != nil {
		panic(err.Error())
	}

	db.SetMaxIdleConns(maxConnections)
	db.SetMaxOpenConns(maxConnections)

	// close mysql connection
	defer db.Close()

	log.Printf("Listen http://localhost:%s", serverPort)
	http.HandleFunc("/", requestHandler)
	err = http.ListenAndServe(*listenAddr, nil)

	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}

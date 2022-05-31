package main

import (
	"browser_history/get_histories"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

const sql_limit = "10"

var webkit_variance = time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)
var firefox_variance = time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
var core_data_variance = time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
var global_histories = get_histories.Get_Histories() //map[string]string

func connect_to_db(_db_path string, sql_query string, _browser string) {
	var sqliteDatabase, _ = sql.Open("sqlite", _db_path) // Open the created SQLite File
	defer sqliteDatabase.Close()                         // Defer Closing the database
	if _browser == "chrome" {
		execute_sql_query(sqliteDatabase, sql_query, "chrome")
	} else if _browser == "safari" {
		execute_sql_query(sqliteDatabase, sql_query, "safari")
	} else if _browser == "edge" {
		execute_sql_query(sqliteDatabase, sql_query, "edge")
	} else if _browser == "firefox" {
		execute_sql_query(sqliteDatabase, sql_query, "firefox")
	}
}

func convert_chrome_time(_in_microseconds int64) time.Time {
	var webkit_time = int64(_in_microseconds/1000000) + webkit_variance.UnixMicro()/1000000
	return time.Unix(webkit_time, 0)
}

func convert_firefox_time(_in_nanoseconds int64) time.Time {
	var firefox_time = int64(_in_nanoseconds/1000000) + firefox_variance.UnixNano()/1000000
	return time.Unix(firefox_time, 0)
}

func convert_core_data_time(_in_nanoseconds float64) time.Time {
	var test = _in_nanoseconds
	fmt.Println(test)
	var core_data_time = int64(_in_nanoseconds) + core_data_variance.Unix()
	return time.Unix(core_data_time, 0)
}

func execute_sql_query(db *sql.DB, _sql_query string, _browser string) {
	rows, err := db.Query(_sql_query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var _url string
	if _browser == "chrome" || _browser == "edge" {
		for rows.Next() { // Iterate and fetch the records from result cursor
			var _last_visit_time int64
			var _visit_count string
			err := rows.Scan(&_url, &_visit_count, &_last_visit_time)
			if err != nil {
				log.Fatal(err)
			}
			webkit_time_converted := convert_chrome_time(int64(_last_visit_time))
			log.Println(_browser, _url, _visit_count, webkit_time_converted)
		}
	} else if _browser == "firefox" {
		for rows.Next() { // Iterate and fetch the records from result cursor
			var _last_visit_time sql.NullInt64
			var _visit_count sql.NullString
			err := rows.Scan(&_url, &_visit_count, &_last_visit_time)
			if err != nil {
				log.Fatal(err)
			}
			firefox_time_converted := convert_firefox_time(_last_visit_time.Int64)
			log.Println(_browser, _url, _visit_count.String, firefox_time_converted)
		}
	} else if _browser == "safari" {
		for rows.Next() { // Iterate and fetch the records from result cursor
			var _item_id string
			var _last_visit_time float64
			var _visit_count string
			err := rows.Scan(&_item_id, &_url, &_visit_count, &_last_visit_time)
			if err != nil {
				log.Fatal(err)
			}
			var _fixed_time = convert_core_data_time(_last_visit_time)
			log.Println(_browser, _url, _visit_count, _fixed_time)
		}
	}
}

func main() {
	for key := range global_histories {
		browser := strings.Split(key, "-")[0]
		if browser == "firefox" {
			var firefox_sql = `SELECT url, visit_count, last_visit_date 
			FROM moz_places 
			ORDER BY last_visit_date DESC LIMIT ` + sql_limit + ";"
			connect_to_db(global_histories[key], firefox_sql, "firefox")
		} else if browser == "safari" {
			var safari_sql = `SELECT history_items.id, history_items.url, history_items.visit_count, history_visits.visit_time
			FROM history_items
			INNER JOIN history_visits on history_items.id=history_visits.history_item 
			ORDER BY history_visits.visit_time DESC LIMIT ` + sql_limit + ";"
			connect_to_db(global_histories["safari"], safari_sql, "safari")
		} else if browser == "edge" {
			var edge_sql = "SELECT url, visit_count, last_visit_time FROM urls ORDER BY last_visit_time DESC LIMIT " + sql_limit + ";"
			connect_to_db(global_histories["edge"], edge_sql, "edge")
		} else if browser == "chrome" {
			var chrome_sql = "SELECT url, visit_count, last_visit_time FROM urls ORDER BY last_visit_time DESC LIMIT " + sql_limit + ";"
			connect_to_db(global_histories["chrome"], chrome_sql, "chrome")
		}
	}
}

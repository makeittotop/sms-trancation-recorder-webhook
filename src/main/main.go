package main

import (
    "fmt"
    "net/http"
    "log"
    //"io/ioutil"
    "strings"
    "database/sql"
    _ "github.com/lib/pq"
)

func setup_db() (db *sql.DB) {
	//postgres_uri := "postgres://ec2-user:test@localhost:8887/twilio?sslmode=disable"
	postgres_uri := "postgres://ec2-user:test@52.6.88.35:5432/twilio?sslmode=disable"
    //db, err := sql.Open("postgres", "user=ec2-user password=test dbname=twilio sslmode=disable")	
    db, err := sql.Open("postgres", postgres_uri)
    
	panic_if(err)
	return	
}

func query_db(db *sql.DB, query string) (rows *sql.Rows) {
	rows, err := db.Query(query)
	panic_if(err)
	return
}

func update_db(db *sql.DB, query string) (rows *sql.Rows) {
	rows, err := db.Query(query)
	panic_if(err)
	return
}

func insert_db(db *sql.DB, data map[string]string) {
	var id int
	
	/*
	keys := make([]string, len(data))
	
	i := 0
	for k := range data {
		keys[i] = k
		i += 1
	}
	*/
	
	err := db.QueryRow(`INSERT INTO sms.data (tonumber, fromnumber,  accountsid,  smssid,  smsstatus,  messagesid,  messagestatus, ts_sent)
	VALUES($1, $2, $3, $4, $5, $6, $7, now()) RETURNING id`, data["tonumber"], data["fromnumber"], data["accountsid"], data["smssid"], data["smsstatus"], data["messagesid"], data["messagestatus"]).Scan(&id)
	panic_if(err)
	fmt.Printf("Succesfully inserted row with id: %d\n", id)
}

func panic_if(err error) {
	if err != nil {
		panic(err)
	}
}

func print_rows(rows *sql.Rows) {
	panic_if(rows.Err())
	for rows.Next() {
		/*
		var id int
		var str string
		if err := rows.Scan(&id, &str); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%d is '%s'\n", id, str)
		*/
		var id, ts int
		var to, from, account_sid, sms_sid, sms_status, message_sid, message_status string
		err := rows.Scan(&id, &ts, &to, &from, &account_sid, &sms_sid, &sms_status, &message_sid, &message_status)
		panic_if(err)
		fmt.Printf("%d, %s, %s, %s, %s, %s, %s, %s\n", id, to, from, account_sid, sms_sid, sms_status, message_sid, message_status); 
	}
}

func main() {
	http.HandleFunc("/twilio", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("==================")
		fmt.Println("Received a req ...")

		// Parse req form
		err := req.ParseForm()
		panic_if(err)
		
		/*
		// Method
		fmt.Printf("Method: %s\n", req.Method)
			
		// Referer
		fmt.Printf("Referer: %s\n", req.Referer());
			
		// Sender
		fmt.Printf("Sender: %s\n", req.RemoteAddr)
			
		// Request Path
		fmt.Printf("Requested URI path: %s\n", req.URL.RequestURI());
			
		// Content length
		fmt.Printf("Content length: %d\n", req.ContentLength)
			
		// Header
		fmt.Printf("Header:\n");
		for k, v := range req.Header {
			fmt.Printf("  %s: %s\n", k, v)
		}
			
		// Body
		body, err := ioutil.ReadAll(req.Body)
		body_str := strings.Trim(string(body), "\n")
		//if body_str != "" {
		fmt.Printf("Body: %s\n", body_str);
		//}
			
		// Form data
		if req.Form != nil {
			fmt.Println("Form data: ");
			for k, v := range req.Form {
			    fmt.Printf("  %s -> %s\n", k, v);
			}
		}
			
		// Post form data
		if req.PostForm != nil {
			fmt.Println("Post form data: ");
			for k, v := range req.PostForm {
			    fmt.Printf("  %s -> %s\n", k, v);	
			}	
		}    	
		*/
		
		// Post form data
		var post_data = make(map[string]string)
		var sms_sid string
		if req.PostForm != nil {
			for k, v := range req.PostForm {
			    //fmt.Printf("  %s -> %s\n", k, v);	
			    k = strings.ToLower(k)
			    if k == "to" {
			    	k = "tonumber"
			    } else if k == "from" {
			    	k = "fromnumber"
			    } else if k == "apiversion" {
			    	continue
			    } else if k == "smssid" {
			    	sms_sid = v[0]
			    }
			    post_data[k] = v[0]
			}	
		}		
		
		db := setup_db()
		defer db.Close()
		
		query := fmt.Sprintf("select * from sms.data where smssid = '%s'",  sms_sid)
		fmt.Println(query)
		rows := query_db(db, query)
	    defer rows.Close()
	    // No data found, insert data in the db
	    if rows.Next() == false {
	    	insert_db(db, post_data)
	    } else {
	    	// Data found - ideally there should only be one entry in 
	    	// that case update the table with the relevant data, namely - smsstatusfinal and messagestatusfinal
	    	query = fmt.Sprintf("update sms.data set smsstatusfinal='%s', messagestatusfinal='%s', ts_delivered=now() where smssid='%s'", post_data["smsstatus"], post_data["messagestatus"], post_data["smssid"])
	    	update_db(db, query)
	    	fmt.Println("table succesfully updated")
	    }
	    //rows := query_db(db, "SELECT * FROM sms.test")
	    /*
		query := fmt.Sprintf("select * from sms.data")
		rows := query_db(db, query)
	    defer rows.Close()
	    print_rows(rows)
	    */
		fmt.Println("==================")

	})
	
	fmt.Println("Starting the web server running on 9999 ...")
	log.Fatal(http.ListenAndServe(":9999", nil))
}

package main

import (
	"io"
	"net/http"
	"fmt"
	"net"
	"time"
	"golang.org/x/net/http2"
	"strconv"
	"bytes"
)

const (
	PORT       = ":443"
	KEY   = "./localhost.key"
	CERT = "./localhost.crt"
)

var conn net.Conn

func DummyData(num int) string {
	var buffer bytes.Buffer
	for i := 0; i < num; i++ {
		buffer.WriteString("x")
	}
	return buffer.String()
}

func hello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hello world! :)")
}

func testCaseMaxFrameSize(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Enter.. \n")

	fr := http2.NewFramer(conn, conn)

	fmt.Printf(strconv.FormatBool(fr.AllowIllegalWrites))

	max_size := 18384

	fr.WriteData(1, true, []byte(DummyData(int(max_size)+1)))
}



func main() {
	http.HandleFunc("/", hello)
	http.HandleFunc("/4.2", testCaseMaxFrameSize);

	s := &http.Server{
		Addr:           PORT,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		ConnState:      ConnStateListener,
	}

	s.ListenAndServeTLS(CERT, KEY)
}

func ConnStateListener(c net.Conn, cs http.ConnState) {



	fmt.Printf("CONN STATE: %v, %v\n", cs, c)

	if cs.String() == "new" {
		conn = c
	}

}

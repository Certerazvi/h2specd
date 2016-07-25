package main

import (
	"io"
	"bytes"
	"time"
	"h2specd"
	"net"
	"fmt"
	"golang.org/x/net/http2/hpack"
)

const (
	SETUP_PORT       = ":443"
	RUNNING_PORT	 = ":1443"
	KEY   		 = "./localhost.key"
	CERT 		 = "./localhost.crt"
	RUN_URL 	 = "https://localhost:1443/RUN_TEST"
)

var conn net.Conn

func switchManager(w h2specd.ResponseWriter, r *h2specd.Request) {
	fmt.Printf("Running test case... ")
	h2specd.RedirectHandler(RUN_URL, h2specd.StatusSeeOther).ServeHTTP(w, r)
	h2specd.ListenerForServer.Close()
}

func pair(name, value string) hpack.HeaderField {
	return hpack.HeaderField{Name: name, Value: value}
}

func DummyData(num int) string {
	var buffer bytes.Buffer
	for i := 0; i < num; i++ {
		buffer.WriteString("x")
	}
	return buffer.String()
}

func runTestCase(w h2specd.ResponseWriter, r *h2specd.Request) {

	if h2specd.TestNo == h2specd.InvalidHeaderTestCase {
		fmt.Fprintf(conn, "\x00\x00\x01\x01\x05\x00\x00\x00\x01\x40")
	}
}

func hello(w h2specd.ResponseWriter, r *h2specd.Request) {
	io.WriteString(w, "Hello world! :)")
}

// 3.5
func testCasePreface(w h2specd.ResponseWriter, r *h2specd.Request) {

	fmt.Printf("Testing if invalid preface is handled correctly: \n")
	h2specd.TestNo = h2specd.PrefaceTestCase
	switchManager(w, r)
}

// 4.3
func testCaseInvalidHeaderBlock(w h2specd.ResponseWriter, r *h2specd.Request) {

	fmt.Printf("Testing how it handles an invalid header: \n")
	h2specd.TestNo = h2specd.InvalidHeaderTestCase
	switchManager(w, r)

}

// 5.1
func testCaseIllegalFrameSentWhileIdle(w h2specd.ResponseWriter,
                                       r *h2specd.Request) {

	fmt.Printf("Testing how it handles an RST_STREAM frame while IDLE: \n")
	h2specd.TestNo = h2specd.IllegalRST_STREAMFrameWhileIdleTestCase
	switchManager(w, r)
}

// 5.3
func testCaseSelfDependingPriorityFrame(w h2specd.ResponseWriter,
					r *h2specd.Request) {

	fmt.Printf("Testing client by sending self depending priority frame: \n")
	h2specd.TestNo = h2specd.SelfDependingPriorityFrameTestCase

	var pp h2specd.Http2PriorityParam
	pp.StreamDep = 2
	pp.Exclusive = false
	pp.Weight = 255

	h2specd.ServerConn.Framer().WritePriority(2, pp)
}

// 5.4
func testCaseGoAwayFrameFollowedByClosedConnection(w h2specd.ResponseWriter,
						   r *h2specd.Request) {

	fmt.Printf("Testing if client closes connection after sending GO AWAY" +
		   " away frame: \n")
	h2specd.TestNo = h2specd.CloseConnAfterGoAwayFrameTestCase
	fmt.Fprintf(conn, "\x00\x00\x08\x06\x00\x00\x00\x00\x03") // PING frame with invalid stream ID
	fmt.Fprintf(conn, "\x00\x00\x00\x00\x00\x00\x00\x00")
	switchManager(w, r)
}

// 5.5
func testCaseDiscardingUnknownFrames(w h2specd.ResponseWriter,
				     r *h2specd.Request) {

	fmt.Printf("Testing if client discards unknown frames: \n")
	h2specd.TestNo = h2specd.DiscardingUnknownFramesTestCase

	// Write a frame of type 0xFF, which isn't yet defined
	// as an extension frame. This should be ignored; no GOAWAY,
	// RST_STREAM or closing the connection should occur
	h2specd.ServerConn.Framer().WriteRawFrame(0xFF, 0x00, 0, []byte("unknown"))

	// Now send a normal PING frame, and if this is processed
	// without error, then the preceeding unknown frame must have
	// been processed and ignored.
	data := [8]byte{'h', '2', 's', 'p', 'e', 'c'}
	h2specd.ServerConn.Framer().WritePing(false, data)
	switchManager(w, r)

}

// 6.1
func testCaseDataFrameWith0x0StreamIndentifier(w h2specd.ResponseWriter,
					       r *h2specd.Request) {

	fmt.Printf("Testing client's response to data frame with 0x0 " +
	 	   "stream identifier: \n")
	h2specd.TestNo = h2specd.DataFrameWith0x0StreamIdentTestCase
	h2specd.ServerConn.Framer().WriteData(0, true, []byte("test"))
	switchManager(w, r)
}

// 6.4.1
func testCaseRST_STREAMFrame0x0Ident(w h2specd.ResponseWriter,
				     r *h2specd.Request) {

	fmt.Printf("Testing if client responds with PROTOCOL_ERROR to a " +
		   "RST_FRAME with 0x0 stream identifier: \n")
	h2specd.TestNo = h2specd.RST_FRAMEWith0x0StreamIdentTestCase
	h2specd.ServerConn.Framer().WriteRSTStream(0, h2specd.Http2ErrCodeCancel)
	switchManager(w, r)

}

// 6.4.2
func testCaseIllegalSizeRST_STREAM(w h2specd.ResponseWriter,
			           r *h2specd.Request) {

	hdrs := []hpack.HeaderField{
		pair(":status", "200"),
		//pair("content-type", "text/html"),
	}

	var hp h2specd.Http2HeadersFrameParam
	var writeBuff bytes.Buffer
	var encoder *hpack.Encoder
	encoder = hpack.NewEncoder(&writeBuff)

	writeBuff.Reset()

	for _, hf := range hdrs {
		_ = encoder.WriteField(hf)
	}

	dst := make([]byte, writeBuff.Len())
	copy(dst, writeBuff.Bytes())


	hp.StreamID = 1
	hp.EndStream = false
	hp.EndHeaders = true
	hp.BlockFragment = dst

	h2specd.ServerConn.Framer().WriteHeaders(hp)

	fmt.Printf("Testing client by sending RST_FRAME with a length other" +
		   "than 4 octets: \n")
	fmt.Fprintf(conn, "\x00\x00\x03\x03\x00\x00\x00\x00\x01")
	fmt.Fprintf(conn, "\x00\x00\x00")
}

// 6.5.1
func testCaseSettingsAck(w h2specd.ResponseWriter, r *h2specd.Request) {

	fmt.Printf("Testing if client sends Settings with ACK after receiving" +
		   " a Setting Frame: \n")

	h2specd.TestNo = h2specd.SettingsACKTestCase

	settings := []h2specd.Http2Setting{
		h2specd.Http2Setting{h2specd.Http2SettingMaxConcurrentStreams, 100},
		h2specd.Http2Setting{h2specd.Http2SettingHeaderTableSize, ^uint32(0)},
	}
	h2specd.ServerConn.Framer().WriteSettings(settings...)
	switchManager(w, r)

}

// 6.5.2
func testCaseNonZeroLengthAckSettingFrame(w h2specd.ResponseWriter,
					  r *h2specd.Request) {

	fmt.Printf("Testing the response to a ACKed Settings Frame with a " +
		   "non-zero length: \n")
	time.Sleep(2 * time.Second)
	fmt.Fprintf(conn, "\x00\x00\x01\x04\x01\x00\x00\x00\x00\x00")

}

// 6.7.1
func testCaseReceivingPingFrame(w h2specd.ResponseWriter, r *h2specd.Request) {

	fmt.Printf("Testing the client by checking the response to a PING " +
	 	   "frame: \n")
	h2specd.TestNo = h2specd.PingFrameReplyTestCase
	data := [8]byte{'h', '2', 's', 'p', 'e', 'c'}
	h2specd.SentData = data
	h2specd.ServerConn.Framer().WritePing(false, data)
	switchManager(w, r)
}

// 6.7.2
func testCasePingWithNonZeroIdent(w h2specd.ResponseWriter,
				  r *h2specd.Request) {

	fmt.Printf("Testing the client by sending a Ping Frame with a " +
	 	   "non-zero stream identifier: \n")
	h2specd.TestNo = h2specd.NonZeroIdentPingFrameTestCase
	fmt.Fprintf(conn, "\x00\x00\x08\x06\x00\x00\x00\x00\x03")
	fmt.Fprintf(conn, "\x00\x00\x00\x00\x00\x00\x00\x00")
	switchManager(w, r)
}

// 6.7.3
func testCasePingWithLengthDiffFromEight(w h2specd.ResponseWriter,
					 r *h2specd.Request) {

	fmt.Printf("Testing client's reaction to a Ping Frame that has the " +
		   "length field different from 8: \n")
	h2specd.TestNo = h2specd.PingFrameWithLengthDiffFromEightTestCase
	fmt.Fprintf(conn, "\x00\x00\x06\x06\x00\x00\x00\x00\x00")
	fmt.Fprintf(conn, "\x00\x00\x00\x00\x00\x00")
	switchManager(w, r)

}

// 6.8
func testCaseGoAwayWithStreamIdentNonZero(w h2specd.ResponseWriter,
					  r *h2specd.Request) {

	fmt.Printf("Testing client by sending a Go Frame with a non zero " +
		   "stream identifier: \n")
	h2specd.TestNo = h2specd.GoAwayWithNonZeroStreamIdentTestCase
	fmt.Fprintf(conn, "\x00\x00\x08\x07\x00\x00\x00\x00\x03")
	fmt.Fprintf(conn, "\x00\x00\x00\x00\x00\x00\x00\x00")
	switchManager(w, r)

}

// 6.9
func testCaseWindowFrameWithZeroFlowControlWindowInc(w h2specd.ResponseWriter,
						     r *h2specd.Request) {

	fmt.Printf("Testing client by sending Window Frame with a flow " +
		   "control window increment of zero: \n")
	h2specd.TestNo = h2specd.ZeroFlowControlWindowIncrementTestCase
	h2specd.ServerConn.Framer().WriteWindowUpdate(0, 0)
	switchManager(w, r)

}

func main() {

	h2specd.HandleFunc("/", hello)
	h2specd.HandleFunc("/3.5", testCasePreface) // checked √
	h2specd.HandleFunc("/4.3", testCaseInvalidHeaderBlock) // checked √
	h2specd.HandleFunc("/5.1", testCaseIllegalFrameSentWhileIdle)
	h2specd.HandleFunc("/5.3", testCaseSelfDependingPriorityFrame)
	h2specd.HandleFunc("/5.4", testCaseGoAwayFrameFollowedByClosedConnection) // checked √
	h2specd.HandleFunc("/5.5", testCaseDiscardingUnknownFrames) // checked √
	h2specd.HandleFunc("/6.1", testCaseDataFrameWith0x0StreamIndentifier) // checked √
	h2specd.HandleFunc("/6.4.1", testCaseRST_STREAMFrame0x0Ident) // checked √
	h2specd.HandleFunc("/6.4.2", testCaseIllegalSizeRST_STREAM)
	h2specd.HandleFunc("/6.5.1", testCaseSettingsAck) // checked √
	h2specd.HandleFunc("/6.5.2", testCaseNonZeroLengthAckSettingFrame)
	h2specd.HandleFunc("/6.7.1", testCaseReceivingPingFrame) // checked √
	h2specd.HandleFunc("/6.7.2", testCasePingWithNonZeroIdent) // checked √
	h2specd.HandleFunc("/6.7.3", testCasePingWithLengthDiffFromEight) // checked √
	h2specd.HandleFunc("/6.8", testCaseGoAwayWithStreamIdentNonZero) // checked √
	h2specd.HandleFunc("/6.9", testCaseWindowFrameWithZeroFlowControlWindowInc) // checked √
	h2specd.HandleFunc("/RUN_TEST", runTestCase)

	s := &h2specd.Server{
		Addr:           SETUP_PORT,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		ConnState:      ConnStateListener,
	}

	s1 := &h2specd.Server{
		Addr:           RUNNING_PORT,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		ConnState:      ConnStateListener2,
	}


	go s1.ListenAndServeTLS(CERT, KEY)

	for {
		s.ListenAndServeTLS(CERT, KEY)
	}
}

func ConnStateListener(c net.Conn, cs h2specd.ConnState) {

	conn = c
	//fmt.Printf(cs.String() + "\n")


}

func ConnStateListener2(c net.Conn, cs h2specd.ConnState) {

	//fmt.Printf(cs.String() + "\n")
}

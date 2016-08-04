package main

import (
	"bytes"
	"time"
	"h2specd"
	"net"
	"fmt"
	"golang.org/x/net/http2/hpack"
	"html/template"
	"flag"
)

var MAIN_PORT string

const (
	SETUP_PORT       = ":443"
	RUNNING_PORT	 = ":1443"
	KEY   		 = "./localhost.key"
	CERT 		 = "./localhost.crt"
	RUN_URL 	 = "https://localhost:1443/RUN_TEST"
)

var conn 			net.Conn
var SetupServer		 	*h2specd.Server
var RunningServer 		*h2specd.Server

func switchToRunningManager(w h2specd.ResponseWriter, r *h2specd.Request) {
	fmt.Printf("Running test case... ")
	//mainTemplate, _ = template.ParseFiles("test_html.tmpl")
	//mainTemplate.Execute(w, nil)
	h2specd.RedirectHandler(RUN_URL, h2specd.StatusSeeOther).ServeHTTP(w, r)
	h2specd.ListenerForTestServer.Close()
}

func getAddress(sectionNo string) string {

	return "https://localhost:443/" + sectionNo

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

	//mainTemplate, _ = template.ParseFiles("test_html.tmpl")
	//mainTemplate.Execute(w, nil)
	if h2specd.TestNo == h2specd.InvalidHeaderTestCase {
		fmt.Fprintf(conn, "\x00\x00\x01\x01\x05\x00\x00\x00\x01\x40")
	}
	//h2specd.RedirectHandler("https://localhost:2443/", h2specd.StatusSeeOther).ServeHTTP(w, r)
}

var mainTemplate *template.Template

func hello(w h2specd.ResponseWriter, r *h2specd.Request) {
	//io.WriteString(w, "Hello world! :)")

	//t := template.New("fieldname example")
	//t, _ = t.Parse("hello {{.UserName}}!")
	//p := Person{UserName: "Astaxie"}
	//t.Execute(w, p)

	//go runTest()
	//
	//time.Sleep(2 * time.Second)

	//s1, _ := template.ParseFiles("header.tmpl", "content.tmpl", "footer.tmpl")
	//s1.ExecuteTemplate(w, "header", nil)
	//s1.ExecuteTemplate(w, "content", nil)
	//s1.ExecuteTemplate(w, "footer", nil)

	mainTemplate, _ = template.ParseFiles("html_file.tmpl")
	mainTemplate.Execute(w, h2specd.Result)

	//fmt.Printf("enter:..\n")
	//time.Sleep(2 * time.Second)
	//h2specd.Redirect(w, r, "https://localhost:443/3.5", h2specd.StatusSeeOther)
	//for !h2specd.Done {}
}

// 3.5
func testCasePreface(w h2specd.ResponseWriter, r *h2specd.Request) {

	fmt.Printf("Testing if invalid preface is handled correctly: \n")
	h2specd.TestNo = h2specd.PrefaceTestCase
	switchToRunningManager(w, r)
}

// 4.3
func testCaseInvalidHeaderBlock(w h2specd.ResponseWriter, r *h2specd.Request) {

	fmt.Printf("Testing how it handles an invalid header: \n")
	h2specd.TestNo = h2specd.InvalidHeaderTestCase
	fmt.Fprintf(conn, "\x00\x00\x01\x01\x05\x00\x00\x00\x01\x40")
	switchToRunningManager(w, r)

}

// 5.1
func testCaseIllegalFrameSentWhileIdle(w h2specd.ResponseWriter,
                                       r *h2specd.Request) {

	fmt.Printf("Testing how it handles an RST_STREAM frame while IDLE: \n")
	h2specd.TestNo = h2specd.IllegalRST_STREAMFrameWhileIdleTestCase
	switchToRunningManager(w, r)
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
		   " frame: \n")
	h2specd.TestNo = h2specd.CloseConnAfterGoAwayFrameTestCase
	fmt.Fprintf(conn, "\x00\x00\x08\x06\x00\x00\x00\x00\x03") // PING frame with invalid stream ID
	fmt.Fprintf(conn, "\x00\x00\x00\x00\x00\x00\x00\x00")
	switchToRunningManager(w, r)
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
	switchToRunningManager(w, r)

}

// 6.1
func testCaseDataFrameWith0x0StreamIndentifier(w h2specd.ResponseWriter,
					       r *h2specd.Request) {

	fmt.Printf("Testing the client's response to data frame with 0x0 " +
	 	   "stream identifier: \n")
	time.Sleep(2 * time.Second)
	h2specd.TestNo = h2specd.DataFrameWith0x0StreamIdentTestCase
	h2specd.ServerConn.Framer().WriteData(0, true, []byte("test"))
	switchToRunningManager(w, r)
}

// 6.4.1
func testCaseRST_STREAMFrame0x0Ident(w h2specd.ResponseWriter,
				     r *h2specd.Request) {

	fmt.Printf("Testing if the client responds with PROTOCOL_ERROR to a " +
		   "RST_FRAME with 0x0 stream identifier: \n")
	h2specd.TestNo = h2specd.RST_FRAMEWith0x0StreamIdentTestCase
	h2specd.ServerConn.Framer().WriteRSTStream(0, h2specd.Http2ErrCodeCancel)
	switchToRunningManager(w, r)

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

	fmt.Printf("Testing if the client sends Settings with ACK after receiving" +
		   " a Setting Frame: \n")

	h2specd.TestNo = h2specd.SettingsACKTestCase

	settings := []h2specd.Http2Setting{
		h2specd.Http2Setting{h2specd.Http2SettingMaxConcurrentStreams, 100},
		h2specd.Http2Setting{h2specd.Http2SettingHeaderTableSize, ^uint32(0)},
	}
	h2specd.ServerConn.Framer().WriteSettings(settings...)
	switchToRunningManager(w, r)

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
	switchToRunningManager(w, r)
}

// 6.7.2
func testCasePingWithNonZeroIdent(w h2specd.ResponseWriter,
				  r *h2specd.Request) {

	fmt.Printf("Testing the client by sending a Ping Frame with a " +
	 	   "non-zero stream identifier: \n")
	h2specd.TestNo = h2specd.NonZeroIdentPingFrameTestCase
	fmt.Fprintf(conn, "\x00\x00\x08\x06\x00\x00\x00\x00\x03")
	fmt.Fprintf(conn, "\x00\x00\x00\x00\x00\x00\x00\x00")
	switchToRunningManager(w, r)
}

// 6.7.3
func testCasePingWithLengthDiffFromEight(w h2specd.ResponseWriter,
					 r *h2specd.Request) {

	fmt.Printf("Testing the client's reaction to a Ping Frame that has the " +
		   "length field different from 8: \n")
	h2specd.TestNo = h2specd.PingFrameWithLengthDiffFromEightTestCase
	fmt.Fprintf(conn, "\x00\x00\x06\x06\x00\x00\x00\x00\x00")
	fmt.Fprintf(conn, "\x00\x00\x00\x00\x00\x00")
	switchToRunningManager(w, r)

}

// 6.8
func testCaseGoAwayWithStreamIdentNonZero(w h2specd.ResponseWriter,
					  r *h2specd.Request) {

	fmt.Printf("Testing the client by sending a Go Frame with a non zero " +
		   "stream identifier: \n")
	h2specd.TestNo = h2specd.GoAwayWithNonZeroStreamIdentTestCase
	fmt.Fprintf(conn, "\x00\x00\x08\x07\x00\x00\x00\x00\x03")
	fmt.Fprintf(conn, "\x00\x00\x00\x00\x00\x00\x00\x00")
	switchToRunningManager(w, r)

}

// 6.9.1
func testCaseWindowFrameWithZeroFlowControlWindowInc(w h2specd.ResponseWriter,
						     r *h2specd.Request) {

	fmt.Printf("Testing the client by sending Window Frame with a flow " +
		   "control window increment of zero: \n")
	h2specd.TestNo = h2specd.ZeroFlowControlWindowIncrementTestCase
	h2specd.ServerConn.Framer().WriteWindowUpdate(0, 0)
	switchToRunningManager(w, r)

}

// 6.9.2 -> Receiving PROTOCOL_ERROR instead of FRAME_SIZE_ERROR.
func testCaseWindowFrameWithWrongLength(w h2specd.ResponseWriter,
					r *h2specd.Request) {

	fmt.Printf("Testing client by sending Window Frame with a length " +
		   "other than a multiple of 4 octets: \n")

	fmt.Fprintf(conn, "\x00\x00\x03\x08\x00\x00\x00\x00\x00")
	fmt.Fprintf(conn, "\x00\x00\x01")

}

// 6.9.3
func testCaseInitialSettingsExceedsMaximumSize(w h2specd.ResponseWriter,
					       r *h2specd.Request) {

	fmt.Printf("Tests the client by sending a " +
		   "SETTINGS_INITIAL_WINDOW_SIZE settings with an exceeded " +
		   "maximum window size value: \n")

	fmt.Fprintf(conn, "\x00\x00\x06\x04\x00\x00\x00\x00\x00")
	fmt.Fprintf(conn, "\x00\x04\x80\x00\x00\x00")



}

func main() {


	portPtr := flag.String("port", "2443", "the port number used for the" +
			       " main page, but excluding '443' and '1443'")

	flag.Parse()

	if *portPtr == "443" || *portPtr == "1443" {
		panic("The port cannot be assigned!")
	}

	MAIN_PORT = ":" + *portPtr

	fmt.Printf("The address for testing is: \n")
	fmt.Printf("https://localhost" + MAIN_PORT + "\n")

	mainMux := h2specd.NewServeMux()

	mainMux.HandleFunc("/", hello)

	testMux := h2specd.NewServeMux()

	testMux.HandleFunc("/3.5", testCasePreface) // checked √
	testMux.HandleFunc("/4.3", testCaseInvalidHeaderBlock) // checked √
	testMux.HandleFunc("/5.1", testCaseIllegalFrameSentWhileIdle)
	testMux.HandleFunc("/5.3", testCaseSelfDependingPriorityFrame)
	testMux.HandleFunc("/5.4", testCaseGoAwayFrameFollowedByClosedConnection) // checked √
	testMux.HandleFunc("/5.5", testCaseDiscardingUnknownFrames) // checked √
	testMux.HandleFunc("/6.1", testCaseDataFrameWith0x0StreamIndentifier) // checked √
	testMux.HandleFunc("/6.4.1", testCaseRST_STREAMFrame0x0Ident) // checked √
	testMux.HandleFunc("/6.4.2", testCaseIllegalSizeRST_STREAM)
	testMux.HandleFunc("/6.5.1", testCaseSettingsAck) // checked √
	testMux.HandleFunc("/6.5.2", testCaseNonZeroLengthAckSettingFrame)
	testMux.HandleFunc("/6.7.1", testCaseReceivingPingFrame) // checked √
	testMux.HandleFunc("/6.7.2", testCasePingWithNonZeroIdent) // checked √
	testMux.HandleFunc("/6.7.3", testCasePingWithLengthDiffFromEight) // checked √
	testMux.HandleFunc("/6.8", testCaseGoAwayWithStreamIdentNonZero) // checked √
	testMux.HandleFunc("/6.9.1", testCaseWindowFrameWithZeroFlowControlWindowInc) // checked √
	testMux.HandleFunc("/6.9.2", testCaseWindowFrameWithWrongLength)
	testMux.HandleFunc("/6.9.3", testCaseInitialSettingsExceedsMaximumSize)
	testMux.HandleFunc("/RUN_TEST", runTestCase)

	mainServer := &h2specd.Server{
		Addr:           MAIN_PORT,
		Handler: 	mainMux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	SetupServer = &h2specd.Server{
		Addr:           SETUP_PORT,
		Handler: 	testMux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		ConnState:      ConnStateListener,
	}

	RunningServer = &h2specd.Server{
		Addr:           RUNNING_PORT,
		Handler: 	testMux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		ConnState:      ConnStateListener2,
	}

	go func() {

		runTest()

		for {

			if h2specd.Done {
				runTest()
			}
		}


	}()
	mainServer.ListenAndServeTLS(CERT, KEY)
}

func runTest() {

	//h2specd.HandleFunc("/auto/3.5", testCasePreface) // checked √
	//h2specd.HandleFunc("/auto/4.3", testCaseInvalidHeaderBlock) // checked √
	//h2specd.HandleFunc("/auto/5.1", testCaseIllegalFrameSentWhileIdle)
	//h2specd.HandleFunc("/auto/5.3", testCaseSelfDependingPriorityFrame)
	//h2specd.HandleFunc("/auto/5.4", testCaseGoAwayFrameFollowedByClosedConnection) // checked √
	//h2specd.HandleFunc("/auto/5.5", testCaseDiscardingUnknownFrames) // checked √
	//h2specd.HandleFunc("/auto/6.1", testCaseDataFrameWith0x0StreamIndentifier) // checked √
	//h2specd.HandleFunc("/auto/6.4.1", testCaseRST_STREAMFrame0x0Ident) // checked √
	//h2specd.HandleFunc("/auto/6.4.2", testCaseIllegalSizeRST_STREAM)
	//h2specd.HandleFunc("/auto/6.5.1", testCaseSettingsAck) // checked √
	//h2specd.HandleFunc("/auto/6.5.2", testCaseNonZeroLengthAckSettingFrame)
	//h2specd.HandleFunc("/auto/6.7.1", testCaseReceivingPingFrame) // checked √
	//h2specd.HandleFunc("/auto/6.7.2", testCasePingWithNonZeroIdent) // checked √
	//h2specd.HandleFunc("/auto/6.7.3", testCasePingWithLengthDiffFromEight) // checked √
	//h2specd.HandleFunc("/auto/6.8", testCaseGoAwayWithStreamIdentNonZero) // checked √
	//h2specd.HandleFunc("/auto/6.9.1", testCaseWindowFrameWithZeroFlowControlWindowInc) // checked √
	//h2specd.HandleFunc("/auto/6.9.2", testCaseWindowFrameWithWrongLength)
	//h2specd.HandleFunc("/auto/RUN_TEST", runTestCase)


	h2specd.Done = false

	go RunningServer.ListenAndServeTLS(CERT, KEY)
	SetupServer.ListenAndServeTLS(CERT, KEY)

	for !h2specd.Done {}
	h2specd.TestNo = h2specd.Default
	conn.Close()
	h2specd.ListenerForRunningServer.Close()
	fmt.Printf("DONE! \n")
}

func ConnStateListener(c net.Conn, cs h2specd.ConnState) {

	conn = c
	//fmt.Printf(cs.String() + "\n")


}

func ConnStateListener2(c net.Conn, cs h2specd.ConnState) {

	//fmt.Printf(cs.String() + "\n")
}

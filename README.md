# h2specd
A testing tool for HTTP2 browser implementation.

### BINARY FILE:

Can be found in the Releases section, with the name “h2specd”.

### INSTRUCTIONS:

a) In order to create an https server on your localhost you need a self-signed certificate, which can be generated by running: “openssl req -x509 -newkey rsa:2048 -keyout localhost.key -out localhost.crt -days XXX -nodes”. You are recommended to run this command in the folder of the binary file so you do not have to move files around.

b) Make sure to run the file with root permissions (this is the case because in order to listen to a port destined for https, root rights are needed).

c) Once the program is running you can access the localhost through your browser and enter “https://localhost:443/XXX” as the address, where XXX represents the section number (based from the RFC 7540 specification) that you want to check in your browser.

d) Each test case will either display PASSED or FAILED in your terminal after doing the above mentioned aspects.

** At the moment it is necessary to force close and restart the binary file in order to test other cases.

### ISSUES:

The source code is based on the net/http golang library, which was modified in certain places (namely server.go and h2_bundle.go) to be able to create tests that focus on extreme cases. Feedback is welcomed on the style, structure and implementation of h2specd. This still represents a work in progress.

Known issues are that the test cases:

testCaseIllegalFrameSentWhileIdle,
testCaseSelfDependingPriorityFrame,
testCaseIllegalSizeRST_STREAM,
testCaseNonZeroLengthAckSettingFrame

do not seem to function as they should when checking them through Wireshark. No GOAWAY frame is received or a different error than the one expected and specified in RFC 7540. These are results from trying the latest version of the Chrome browser.

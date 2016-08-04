# h2specd
A testing tool for HTTP2 browser implementation.

### BINARY FILE:

Can be found in the [Releases section](https://github.com/Certerazvi/h2specd/releases), with the name "h2specd” and followed by the OS and architecture it is destined for. Download the one suitable for you.

### INSTRUCTIONS:

**a)** In order to create an https server on your localhost you need a self-signed certificate, which can be generated. You are recommended to run this command in the folder of the binary file so you do not have to move files around.

    % openssl req -x509 -newkey rsa:2048 -keyout localhost.key -out localhost.crt -days <DAYS> -nodes

**b)** Currently the only option available is setting a different port number to use for the main page. The default is 2443 hence the binary file can be run as it this, through the terminal. However, be careful to **NOT** use 443 or 1443 as a custom port for this. They are already in use within the program. Moreover, make sure to run the file with root permissions (this is the case because in order to listen to a port destined for https, root rights are needed). For Linux and MacOS, type:

    % sudo ./h2specd
    or
    % sudo ./h2specd -port <PORT>

**c)** Once the program is running you can access the localhost through your browser and enter “https://localhost:<PORT>/” as the address, where the PORT is the one you chose or 2443 by default. Following this, links to different tests will be displayed along with short descriptions. You can either click on one of them to individually test it or you can click the "Auto Test" button to run all the tests (with the exception of the first one which still has a few issues).

    https://localhost:<PORT>/

**d)** After following the above, within the terminal, information about the outcome of the test will appear. Moreover, in the browser, after autotesting, ticks should be displayed in front of the tests that passed and crosses in front of the ones that did not.

\*\* In case that a test blocks while running, it is recommended to force close and restart the binary file.

### ISSUES:

The source code is based on the net/http golang library, which was modified in certain places (namely server.go and h2_bundle.go) to be able to create tests that focus on extreme cases. Feedback is welcomed on the style, structure and implementation of h2specd. This still represents a work in progress.

Known issues are that the test cases:

* testCaseIllegalFrameSentWhileIdle
* testCaseSelfDependingPriorityFrame
* testCaseIllegalSizeRST_STREAM
* testCaseNonZeroLengthAckSettingFrame

do not seem to function as they should when checking them through Wireshark. No GOAWAY frame or a different error than the one expected and specified in RFC 7540 is received. These are results from trying the latest version of the Firefox browser.

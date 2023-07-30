package httpserver

import (
	"bufio"
	"log"
	"net"
	"strings"
)

import (
	"bytes"
)

type HttpInfo struct {
	host   string
	method string
	path   string
}

func (http HttpInfo) Host() string {
	return http.host
}

func extractHttpInfo(conn net.Conn) (*HttpInfo, *bufio.Reader, []byte) {
	var info HttpInfo
	r := bufio.NewReader(conn)

	firstLine := true
	var buff bytes.Buffer
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			log.Println(err)
			break
		}
		buff.WriteString(msg)
		if firstLine {
			fl := strings.Split(msg, " ")
			info.method = fl[0]
			info.path = fl[1]

		}
		firstLine = false
		if strings.Contains(msg, "Host: ") {
			info.host = msg[6 : len(msg)-2]
			break
		}
	}

	return &info, r, buff.Bytes()
}

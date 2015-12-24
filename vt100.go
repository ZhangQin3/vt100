package vt100

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net"
	"time"
)

type Vt100 struct {
	addr string
	conn net.Conn
	bio  *bufio.Reader
}

const buferSize = 4096

func Connect(ip string) (Vt100, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ip+":telnet")
	if err != nil {
		return Vt100{}, err
	}

	conn, err := net.DialTCP("tcp4", nil, tcpAddr)
	vt := Vt100{ip, conn, bufio.NewReaderSize(conn, buferSize)}

	return vt, err
}

func (v *Vt100) Close() {
	v.conn.Close()
}

func (v *Vt100) Recv(timeout time.Duration) {
	var b bytes.Buffer
	data := make([]byte, 512)
	end := time.Now().Add(timeout * time.Second)

	for {
		v.conn.SetReadDeadline(time.Now().Add(1000 * time.Millisecond))
		n, err := v.bio.Read(data)

		b.Write(data[:n])

		fmt.Println(n, "==> ", string(data[:n]))

		if err != nil || time.Now().After(end) {
			break
		}
	}
}

func (v *Vt100) setReadTimeout(timeout time.Duration) {
	v.conn.SetReadDeadline(time.Now().Add(timeout * time.Millisecond))
}

func (v *Vt100) RecvBytes(delim byte, timeout time.Duration) ([]byte, error) {
	var b bytes.Buffer

	v.conn.SetReadDeadline(time.Now().Add(timeout * time.Second))
	buf, err := v.bio.ReadBytes(delim)
	b.Write(buf)

	fmt.Println("==> ", string(buf), err)
	return b.Bytes(), err
}

func (v *Vt100) RecvUntil(until string, timeout time.Duration) (res []byte, err error) {
	var end = time.Now().Add(timeout * time.Second)
	var u = []byte(until)

	for {
		// the timeout here should be smaller, but digatto's first screen take more time than 1s sometimes.
		v.setReadTimeout(100)

		b, err := v.bio.Peek(buferSize)
		// fmt.Println("==> ", string(b), err)

		if index := bytes.LastIndex(b, u); index != -1 {
			l := index + len(u)
			res = make([]byte, l)
			copy(res, b[:l])
			v.bio.Discard(l)
			return res, nil
		}

		if e, ok := err.(*net.OpError); ok && e.Timeout() {
			err = nil
		}

		if time.Now().After(end) {
			err = errors.New("vt100 timeout")
		}

		if err != nil {
			res = make([]byte, len(b))
			copy(res, b)
			v.bio.Discard(len(b))
			return res, err
		}
	}
}

func (v *Vt100) RecvAtLeast(subStr string, timeout time.Duration) (res []byte, err error) {
	var end = time.Now().Add(timeout * time.Second)
	var sub = []byte(subStr)

	for {
		// the timeout here should be smaller, but digatto's first screen take more time than 1s sometimes.
		v.setReadTimeout(100)

		b, err := v.bio.Peek(buferSize)
		// fmt.Println("==> ", string(b), err)

		if ok := bytes.Contains(b, sub); ok {
			res = make([]byte, v.bio.Buffered())
			v.bio.Read(res)
			return res, nil
		}

		if e, ok := err.(*net.OpError); ok && e.Timeout() {
			err = nil
		}

		if time.Now().After(end) {
			err = errors.New("vt100 timeout")
		}

		if err != nil {
			res = make([]byte, len(b))
			v.bio.Read(res)
			return res, err
		}
	}
}

func (v *Vt100) CursorDown() {
	var b = []byte{'\x1B', '[', 'B'}
	v.conn.Write(b)
}

func (v *Vt100) SendEnter() {
	var b = []byte{'\r', '\x00'}
	v.conn.Write(b)
}

func (v *Vt100) Send(cmd string) {
	var b = []byte(cmd)
	v.conn.Write(b)
	// between two write options must inserts a time interval, otherwith the second writing option may not work.
	time.Sleep(10 * time.Millisecond)
}

// Send A digit from 0 to 9
func (v *Vt100) sendDigit(i int) {
	var b = []byte{byte(48 + i)}
	v.conn.Write(b)
	// between two write options must inserts a time interval, otherwith the second writing option may not work.
	time.Sleep(10 * time.Millisecond)
}

// Send the port number, the number i must less 100.
func (v *Vt100) SendPortNum(i int) {
	v.sendDigit(i / 10)
	v.sendDigit(i % 10)
}

func (v *Vt100) MoveOnTo(value string) {
	for i := 0; i < 20; i++ {
		v.CursorDown()
		b, err := v.RecvUntil("[0;10m", 3)
		fmt.Println("==> ", string(b), err)

		if bytes.Contains(b, []byte(value)) {
			break
		}
	}
}

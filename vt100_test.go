package vt100

import (
	"fmt"
	"testing"
)

func TestVt100(t *testing.T) {
	conn, err := Connect("10.89.255.104")
	fmt.Println(conn, err)

	b, err := conn.RecvUntil("quit this config menu", 5)
	fmt.Println(string(b), err)

	conn.CursorDown()
	b, err = conn.RecvUntil("config port parameter in port", 5)
	fmt.Println(string(b), err)

	conn.SendEnter()
	b, err = conn.RecvUntil("select port to config", 5)
	fmt.Println(string(b), err)

	conn.SendPortNum(22)
	b, err = conn.RecvUntil("2", 5)
	fmt.Println("---------->", string(b), err)

	conn.SendEnter()
	b, err = conn.RecvUntil("Baud of the serial port, from 300 to 460800 Bps", 5)
	fmt.Println("---------->", string(b), err)

	conn.SendEnter()
	b, err = conn.RecvUntil("[19;33H", 5)
	fmt.Println("---------->", string(b), err)

	conn.MoveOnTo("19200")
}

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
)

type Client struct {
	conn     net.Conn
	port     int
	priority int
	leader   bool
}

func sendInfo(curr_con net.Conn) {

	for _, cl := range clientsInfo {
		if curr_con != cl.conn {
			io.WriteString(curr_con, strconv.Itoa(cl.priority)+":"+strconv.Itoa(cl.port)+":"+strconv.FormatBool(cl.leader)+"\n")

		}

	}
	io.WriteString(curr_con, "OK\n")

}

var clientsInfo = make(map[net.Conn]Client)
var max_process int = 5
var priority = 2

func handleConnection(conn net.Conn, clch chan string, pid int) {

	reader := bufio.NewReader(conn)
	for {
		_, err := reader.ReadString('\n')
		if err != nil {
			clch <- "process with id : " + strconv.Itoa(pid) + " has left"
			return
		}
	}
}

//Directory server
func main() {
	clch := make(chan string)
	go printall(clch)
	listen, err := net.Listen("tcp", "127.0.0.1:9008")
	leader := false
	if err != nil {
		log.Fatal(err)
	}
	for ; max_process > 0; max_process = max_process - 1 {
		conn, err := listen.Accept()

		if err != nil {
			fmt.Println("Error while accepting a new connection", err)
			//		continue //move to the next iteration
		}
		//client's impo
		if addr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
			//fmt.Println(addr.Port)

			if max_process == 1 {
				leader = true

			}
			check := false
			var p int
			for check == false {
				p = addr.Port + rand.Intn(5-1) + 1

				ln, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(p))
				if err != nil {
					continue
				}
				check = true
				_ = ln.Close()
			}
			clientsInfo[conn] = Client{conn: conn, priority: priority, port: p, leader: leader}
			clch <- "Process with ID : " + strconv.Itoa(priority) + " has joined"
			io.WriteString(conn, strconv.Itoa(clientsInfo[conn].priority)+":"+strconv.Itoa(clientsInfo[conn].port)+":"+strconv.FormatBool(leader)+"\n")
			go handleConnection(conn, clch, priority)
			priority += 1
		}

	}
	//listen.Close()
	fmt.Println("All connected")
	//var clientChan chan string
	//for msg := range clientChan {
	//	fmt.Printf(msg)
	//}
	for con, _ := range clientsInfo {
		go sendInfo(con)

	}
	var useless int
	fmt.Scan(&useless)
}
func printall(clch chan string) {
	for {

		select {
		case cl := <-clch:
			{
				fmt.Println(cl)
			}
		}
	}
}

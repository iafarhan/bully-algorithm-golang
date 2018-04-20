package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

var me = Client{}
var clientsInfo = make(map[int]*Client)
var serverCons = make(map[int]net.Conn)
var coordinator int

type Client struct {
	s_conn net.Conn
	pid    int
	port   int
	leader bool
}

//Directory server
func readFromServer(conn net.Conn, ch chan<- string, blockCh chan bool, connectCh chan bool) {
	reader := bufio.NewReader(conn)

	for {

		message, err := reader.ReadString('\n')
		//conn.Read(recvdSlice)
		if message != "OK\n" {
			if err != nil {
				break
			}
			s := strings.Split(message, ":")
			pid, _ := strconv.Atoi(s[0])
			port, _ := strconv.Atoi(s[1])
			leader, _ := strconv.ParseBool(s[2][:len(s[2])-1])
			if leader == true {
				coordinator = pid
				fmt.Println("coordinator id  : ", coordinator)

			}
			if (Client{}) == me {
				me.port = port
				me.pid = pid
				me.leader = leader
				blockCh <- true

			} else {
				clientsInfo[pid] = &Client{port: port, leader: leader, pid: pid}

			}
			//	ch <- fmt.Sprintf("%s", string(recvdSlice))

		} else {
			fmt.Println("OK")
			connectOthers()
		}

	}

}

func printRoutine(serverChan chan string) {
	for {
		select {
		case msg := <-serverChan:
			{
				fmt.Println(msg)
			}
		}
	}
}
func serve(ch chan bool, clientCh chan string, rmCh chan int, allCh chan string) {
	_ = <-ch
	fmt.Println("my Process id : ", me.pid)

	fmt.Println("Listening at port ", strconv.Itoa(me.port))

	listen, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(me.port))
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listen.Accept()

		if err != nil {
			fmt.Println("Error while accepting a new connection", err)
			//		continue //move to the next iteration
		}

		go handleConnection(conn, clientCh, rmCh, allCh)

	}
}
func getleaderId() int {
	max := me.pid
	for id, _ := range clientsInfo {
		if id > max {
			max = id
		}
	}
	return max
}
func randomDetector() int {
	slice := make([]int, len(clientsInfo))
	itr := 0
	for pid, _ := range clientsInfo {
		slice[itr] = pid
		itr += 1
	}
	ind := slice[rand.Intn(len(slice))]
	return clientsInfo[ind].pid
}

var detectionMsg string = " "

func connectOthers() {
	for id, cl := range clientsInfo {
		con, err := net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(cl.port))
		if err != nil {
			log.Fatal(err)
		}
		clientsInfo[id].s_conn = con

		go serverConnection(clientsInfo[id].s_conn)

	}
}

func checkVictory() bool {
	return me.pid >= getleaderId()
}
func startElection() {

	if checkVictory() {
		for pid, _ := range clientsInfo {
			leader = true
			detectionMsg = " "
			io.WriteString(serverCons[pid], "c"+strconv.Itoa(me.pid)+"\n")
		}
		fmt.Println("I am the Coordinator ")

	} else {
		for pid, _ := range clientsInfo {
			if pid > me.pid {
				sendCount += 1
				io.WriteString(serverCons[pid], "e"+strconv.Itoa(me.pid)+"\n")
			}
		}
	}
	time.Sleep(time.Duration(5) * time.Microsecond * 100000)
	if okCount == 0 {
		for pid, _ := range clientsInfo {
			leader = true
			detectionMsg = " "
			io.WriteString(serverCons[pid], "c"+strconv.Itoa(me.pid)+"\n")

		}
		fmt.Println("I am the Coordinator ")

	}

}
func handleConnection(conn net.Conn, clientCh chan string, rmCh chan int, allCh chan string) {
	reader := bufio.NewReader(conn)
	id, _ := reader.ReadString('\n')
	str_id := id[:len(id)-1]
	clientCh <- str_id
	pid, _ := strconv.Atoi(str_id)
	serverCons[pid] = conn

	//	clientsInfo[pid].c_conn = conn.(*net.TCPConn)
	defer func() {

		//allCh <- "h"
		if pid == getleaderId() {
			rmCh <- pid
			rand.Seed(time.Now().UnixNano())
			r := rand.Intn(6) + 1

			fmt.Printf("r is : %d\n", r)
			time.Sleep(time.Duration(r) * time.Microsecond * 10000)
			if detectionMsg == " " && leader == true {
				s := "d" + strconv.Itoa(me.pid)

				for _, con := range serverCons {

					io.WriteString(con, s+"\n")

				}
				leader = false
				go startElection()

			}
		} else {
			rmCh <- pid
		}

		//if me.pid == randomDetector() {

		//}

	}()
	for {
		_, err := reader.ReadString('\n')
		if err != nil {
			return
		}
	}
	defer conn.Close()

}

var okCount = 0
var sendCount = 0

func serverConnection(conn net.Conn) {
	io.WriteString(conn, strconv.Itoa(me.pid)+"\n")
	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')
		if len(msg) > 0 {
			fmt.Println()
			if string(msg[0]) == "d" {
				fmt.Printf("process with pid " + msg[1:len(msg)-1] + " Detected Abnormality.\n")
				detectionMsg = "d"
			} else if string(msg[0]) == "c" {

				fmt.Printf("New coordinator is : PID " + msg[1:len(msg)-1] + "\n")
				leader = true
				detectionMsg = " "

			} else if string(msg[0]) == "e" {
				fmt.Printf("Election request by PID : " + msg[1:len(msg)-1] + "\n")
				id, _ := strconv.Atoi(msg[1 : len(msg)-1])
				io.WriteString(serverCons[id], "o"+"\n")
				startElection()
			} else if string(msg[0]) == "o" {
				fmt.Printf("PID  : " + msg[1:len(msg)-1] + " STATUS - OK" + "\n")
				okCount += 1
			}

		}
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func main() {
	conn, err := net.Dial("tcp", ":9008")
	if err != nil {
		fmt.Println("server not found")
		return
	}
	serverChan := make(chan string)

	blockCh := make(chan bool)
	connectCh := make(chan bool)
	clientCh := make(chan string)
	rmCh := make(chan int)
	allCh := make(chan string)
	go readFromServer(conn, serverChan, blockCh, connectCh)
	go handleClients(clientCh, rmCh, allCh)
	go serve(blockCh, clientCh, rmCh, allCh)
	var input int
	fmt.Scanf("%d", &input)
}

var leader bool = true

func handleClients(clientCh chan string, rmCh chan int, allCh chan string) {

	for {
		select {

		case client := <-clientCh:
			{
				fmt.Printf("PID %s is now connected\n", client)

			}
		case client := <-allCh:
			{

				for _, con := range serverCons {

					io.WriteString(con, client+"\n")
				}
			}
		case client := <-rmCh:
			delete(clientsInfo, client)
			delete(serverCons, client)
			//	fmt.Print(client)
			printConnected()
		}

	}
}
func printConnected() {
	println("--Connected Processes -- ")
	for pid, _ := range clientsInfo {

		fmt.Printf("\tPID: %d\n", pid)
	}

}

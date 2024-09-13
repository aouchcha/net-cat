package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type NC struct {
	mutex          sync.Mutex
	Members        map[net.Conn]string
	Member_Numbers int
}

var Archive string

func main() {
	var nc NC = NC{Members: make(map[net.Conn]string)}
	// This function check the command entered by the user if is it valid .
	port := CheckCommand()

	// Creat a server with tcp protocol and a choosen port
	listner, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("We can't creat the server check the code")
		return
	}
	fmt.Println("Listening on the port :" + port)

	for {
		// Accept every connection in an enfinit loop
		connection, err := listner.Accept()
		if err != nil {
			fmt.Println("Error2")
			return
		}
		// Check every connection and handle it
		go HandelConnect(connection, &nc)

	}
}

func CheckCommand() string {
	var port string
	if len(os.Args[1:]) > 1 {
		fmt.Println("[USAGE]: ./TCPChat $port")
	} else if len(os.Args[1:]) == 1 {
		temp, err := strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Println("You enterd a port that maybe containe a caracter !!!")
			os.Exit(1)
		}
		if temp < 1024 || temp > 49151 {
			fmt.Println("You didn't enter a valid port !!! 1024 < port < 49151")
			os.Exit(1)
		}
		port = os.Args[1]
	} else {
		port = "8989"
	}
	return port
}

func Checklettersunder32(users map[net.Conn]string, s []byte) bool {
	for _, char := range s {
		if char < 32 {
			return true
		}
	}
	for _, name := range users {
		if name == string(s) {
			return true
		}
	}

	return false
}

func CheckNames(connection net.Conn, users map[net.Conn]string) []byte {
	var err1 error
	name := make([]byte, 500)
	connection.Write([]byte("Enter your name:"))
	name_len, err := connection.Read(name)
	if err != nil {
		return nil
	}
	temp := name[:name_len-1]

	fmt.Println(string(temp), name_len)
	if name_len-1 == 0 {
		message := "You didn't enter any thing join the server again \n"
		connection.Write([]byte(message))
		err1 = errors.New("empty input")

	} else if Checklettersunder32(users, temp) {
		message := "The name is alreay used choose another one and reconnect to our server\n"
		connection.Write([]byte(message))
		err1 = errors.New("user allridy exist")

	}
	if err1 != nil {
		temp = CheckNames(connection, users)
	}
	fmt.Println(string(temp),name_len)
	return temp
}

func HandelConnect(connection net.Conn, nc *NC) {
	PrintPingouin(connection)

	name := CheckNames(connection, nc.Members)
	if name == nil {
		connection.Close()
		return
	}

	if len(nc.Members) < 10 {
		IsHeIn := true
		LogIOUser(nc.Members, name, IsHeIn)
		nc.mutex.Lock()
		nc.Members[connection] = string(name)
		connection.Write([]byte(Archive))
		nc.mutex.Unlock()

		for {
			text := make([]byte, 10000)
			PrintTimeFormat(nc.Members)
			text_len, err := connection.Read(text)
			if err != nil {
				IsHeIn = false
				temp := name
				nc.mutex.Lock()
				delete(nc.Members, connection)
				nc.mutex.Unlock()
				LogIOUser(nc.Members, temp, IsHeIn)
				PrintTimeFormat(nc.Members)
				connection.Close()
				return
			}
			if Checklettersunder32(nil, text[:text_len-1]) {
				for connect := range nc.Members{
					if connect.RemoteAddr() != connection.RemoteAddr() {
						connect.Write([]byte("\n"))
					}
				}
				continue
			} else {
				PrintForUsers(text[:text_len-1], connection.RemoteAddr().String(), nc.Members, (name), &nc.mutex)
			}
		}
	} else {
		connection.Write([]byte("we are full right now\n"))
		err := connection.Close()
		if err != nil {
			fmt.Println("Error in closing connection")
			return
		}
	}
}

func PrintTimeFormat(users map[net.Conn]string) {
	timeNow := time.Now().Format("2006-01-02 15:04:05")
	for connection := range users {
		format_final := fmt.Sprintf("[%s][%s]: ", timeNow, users[connection])
		connection.Write([]byte(format_final))
	}
}

func PrintForUsers(text []byte, remote string, users map[net.Conn]string, name []byte, mutex *sync.Mutex) {
	var format_final string
	time := time.Now().Format("2006-01-02 15:04:05")
	for rm := range users {
		format_final = fmt.Sprintf("\n[%s][%s]: %s\n", time, name, text)
		if remote != rm.RemoteAddr().String() {
			rm.Write([]byte(format_final))
		} else {
			// fmt.Println(text)
			format_final = fmt.Sprintf("[%s][%s]: %s\n", time, name, text)
			mutex.Lock()
			Archive += format_final
			mutex.Unlock()
			err := os.WriteFile("archive.txt", []byte(Archive), 0o666)
			if err != nil {
				fmt.Println("We can't write into the archive !!!")
				os.Exit(1)
			}
		}
	}
}

func LogIOUser(users map[net.Conn]string, name []byte, IsHEIn bool) {
	var message string
	if IsHEIn {
		message = "\n" + string(name) + " has join our chat...\n"
	} else {
		message = "\n" + string(name) + " has left our chat...\n"
	}
	for connection := range users {
		connection.Write([]byte(message))
	}
}

func PrintPingouin(connection net.Conn) {
	pingouin, err := os.ReadFile("pingouin.txt")
	if err != nil {
		fmt.Println("Error in printing logo")
		os.Exit(1)
	}
	connection.Write(pingouin)
}

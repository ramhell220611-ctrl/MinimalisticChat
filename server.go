package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type server struct {
	mu                sync.RWMutex
	activeConnections map[net.Conn]bool
	users             map[net.Conn]string
	usersIP           map[string]net.Conn
	//
	roomsLmao map[string]string // like map name (#global) & value (slice of users)
}

func main() {

	// context init
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s := &server{
		activeConnections: make(map[net.Conn]bool),
		users:             make(map[net.Conn]string),
		usersIP:           make(map[string]net.Conn),
		roomsLmao:         make(map[string]string), // k = nickname, v = room name
	}

	listener, err := net.Listen("tcp", ":59999")
	if err != nil {
		log.Fatalln("!ERROR!ERROR!ERROR!", err, time.Now().Format("15:04:05 02.01.2006"))
	}
	defer listener.Close()

	go func() {
		<-ctx.Done()
		log.Printf("Server shuts down\n")
		listener.Close()
	}()

	log.Printf("Chat runs without any troubles")

	for {
		c, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				log.Printf("Server has stopped successfully\n")
				return
			default:
				log.Printf("Accept error\n")
				continue
			}
		}

		go s.newConn(c)
	}
}

func (s *server) newConn(conn net.Conn) {
	var delNick string

	s.mu.Lock()
	s.activeConnections[conn] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delNick = s.users[conn]
		delete(s.usersIP, delNick)
		delete(s.users, conn)
		delete(s.activeConnections, conn)
		s.mu.Unlock()
		conn.Close()
	}()

	for {
		buffer := make([]byte, 8192) // 8kb buf idk

		n, err := conn.Read(buffer)
		if err != nil {
			log.Println(err)
			return
		}

		message := strings.TrimSpace(string(buffer[:n]))

		switch {
		case strings.HasPrefix(message, "/who"):
			s.mu.RLock()
			for _, v := range s.users {
				_, err := conn.Write([]byte(v))
				if err != nil {
					log.Println(err)
				}
			}
			s.mu.RUnlock()
		case strings.HasPrefix(message, "/cr"):
			idk := strings.SplitN(message, " ", 2)
			s.mu.Lock()
			actingName := s.users[conn]
			s.roomsLmao[actingName] = idk[1]
			_, err := conn.Write([]byte("console: room has created"))
			if err != nil {
				log.Println(err)
			}
			s.mu.Unlock()
		case strings.HasPrefix(message, "/join"):
			// initing
			idk := strings.SplitN(message, " ", 2)
			// smart way idk (no, i think this can be better)
			s.mu.Lock()
			activeName := s.users[conn]
			delete(s.roomsLmao, activeName)
			roomName := idk[1]
			s.roomsLmao[activeName] = roomName
			conn.Write([]byte("/// " + roomName))
			s.mu.Unlock()
		case strings.HasPrefix(message, "/kkk12111dddqqqprstlon"):
			data := strings.SplitN(message, " ", 3)

			name := data[1]
			room := data[2]

			s.mu.Lock()
			s.users[conn] = name
			s.usersIP[name] = conn
			s.roomsLmao[name] = room
			s.mu.Unlock()
		case strings.HasPrefix(message, "/msg"):
			res := strings.SplitN(message, " ", 3) // 0 = skip, 1 = targetNick, 2 = message
			nick := res[1]
			msg := res[2]

			s.mu.Lock()
			senderNick := s.users[conn]
			targetConn := s.usersIP[nick]
			_, err = targetConn.Write([]byte("*IN PRIVATE* by " + senderNick + ": " + msg))
			s.mu.Unlock()
		default:
			log.Printf("New message: %s", message)
			s.broadcast(conn, message)
		}
	}
}

func (s *server) broadcast(conn net.Conn, entryMessage string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	myNick := s.users[conn]
	myRoom := s.roomsLmao[myNick]

	for clientConn := range s.activeConnections {
		if clientConn == conn {
			continue 
		}
		
		clientNick := s.users[clientConn]
		clientRoom := s.roomsLmao[clientNick]

		if clientRoom == myRoom {
			_, err := clientConn.Write([]byte(entryMessage))
			if err != nil {
				log.Printf("Write error: %v", err)
			}
		}
	}
}

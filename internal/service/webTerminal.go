package service

import (
	"dockernas/internal/backend/docker"
	"dockernas/internal/models"
	"io"
	"log"

	"github.com/gorilla/websocket"
)

func ProcessWebsocketConn(conn *websocket.Conn, instanceName string, rows string, columns string) {
	instance := models.GetInstanceByName(instanceName)
	if instance == nil {
		panic("no instance with name " + instanceName)
	}

	hr := docker.Exec(instance.ContainerID, "100", columns)

	// 退出进程
	defer func() {
		hr.Conn.Write([]byte("exit\r"))
		defer hr.Close()
	}()

	log.Println("websocket attach " + instanceName)

	// 转发输入/输出至websocket
	go func() {
		wsWriterCopy(hr.Conn, conn)
	}()

	wsReaderCopy(conn, hr.Conn)

	log.Println("websocket disattach " + instanceName)
}

// 将终端的输出转发到前端
func wsWriterCopy(reader io.Reader, writer *websocket.Conn) {
	buf := make([]byte, 8192)
	for {
		nr, err := reader.Read(buf)
		if nr > 0 {
			err := writer.WriteMessage(websocket.BinaryMessage, buf[0:nr])
			if err != nil {
				return
			}
		}
		if err != nil {
			return
		}
	}
}

// 将前端的输入转发到终端
func wsReaderCopy(reader *websocket.Conn, writer io.Writer) {
	for {
		messageType, p, err := reader.ReadMessage()
		if err != nil {
			return
		}
		if messageType == websocket.TextMessage {
			writer.Write(p)
		}
	}
}

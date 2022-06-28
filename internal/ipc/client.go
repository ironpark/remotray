package ipc

import (
	"encoding/json"
	"errors"
	gipc "github.com/james-barrow/golang-ipc"
)
import "github.com/gofrs/uuid"

type Client struct {
	ipc           *gipc.Client
	msgChannel    map[string]chan []byte
	errMsgChannel map[string]chan []byte
	eventCallback EventCallback
}

type Msg struct {
	Id          string          `json:"id"`
	MessageType int             `json:"msgType"`
	Data        json.RawMessage `json:"data,omitempty"`
}
type EventCallback func(eventId int, data []byte)

func (c *Client) OnEvent(callback EventCallback) {
	c.eventCallback = callback
}

func (c *Client) ReadReplyMessage(msgId string, dst interface{}) (err error) {
	if c.msgChannel[msgId] == nil {
		return nil
	}
	defer func(msgId string) {
		close(c.msgChannel[msgId])
		close(c.errMsgChannel[msgId])
		delete(c.msgChannel, msgId)
		delete(c.errMsgChannel, msgId)
	}(msgId)

	select {
	case msgData := <-c.msgChannel[msgId]:
		return json.Unmarshal(msgData, dst)
	case err := <-c.errMsgChannel[msgId]:
		return errors.New(string(err))
	}
}

func (c *Client) WriteMessage(msgType int, data interface{}) (msgId string, err error) {
	msgData, _ := json.Marshal(data)
	msg := Msg{
		MessageType: msgType,
		Id:          uuid.Must(uuid.NewV4()).String(),
		Data:        msgData,
	}
	c.msgChannel[msg.Id] = make(chan []byte)
	c.errMsgChannel[msg.Id] = make(chan []byte)

	marshaledData, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}
	err = c.ipc.Write(2, marshaledData)
	if err != nil {
		return "", err
	}
	return msg.Id, nil
}

func NewClient(ipcName string) (*Client, error) {
	client, err := gipc.StartClient(ipcName, &gipc.ClientConfig{
		Timeout:    0,
		RetryTimer: 0,
		Encryption: false,
	})

	if err != nil {
		return nil, err
	}
	// for ipc client ready
	// TODO: check ipc state,err
	client.Read() // Connecting
	client.Read() // Connected
	ipcClient := &Client{
		ipc:           client,
		msgChannel:    map[string]chan []byte{},
		errMsgChannel: map[string]chan []byte{},
	}
	go func(client *Client) {
		for {
			msg, err := ipcClient.ipc.Read()
			if err != nil {
				break
			}
			jmsg := Msg{}
			json.Unmarshal(msg.Data, &jmsg)
			if msg.MsgType == 2 {
				ipcClient.msgChannel[jmsg.Id] <- jmsg.Data
				continue
			}
			if msg.MsgType == -1 {
				ipcClient.errMsgChannel[jmsg.Id] <- jmsg.Data
				continue
			}
			if ipcClient.eventCallback != nil {
				ipcClient.eventCallback(jmsg.MessageType, jmsg.Data)
			}
		}
	}(ipcClient)
	return ipcClient, nil
}

package ipc

import (
	"encoding/json"
	gipc "github.com/james-barrow/golang-ipc"
)

const IpcMsgTypeRaw = 1
const IpcMsgTypeJson = 2
const IpcMsgTypeEvent = 3

type Process func(data []byte) (interface{}, error)
type Server struct {
	ipc                *gipc.Server
	processingCallback map[int]Process
}

func NewServer(ipcName string) (*Server, error) {
	config := &gipc.ServerConfig{UnmaskPermissions: true, Encryption: false}
	sc, err := gipc.StartServer(ipcName, config)
	if err != nil {
		return nil, err
	}
	return &Server{ipc: sc, processingCallback: map[int]Process{}}, nil
}

func (s *Server) SetMessageProcessor(messageType int, process Process) {
	s.processingCallback[messageType] = process
}

func (s *Server) EventEmit(eventType int, data interface{}) {
	innerData, _ := json.Marshal(data)
	msgData, _ := json.Marshal(Msg{
		Id:          "",
		MessageType: eventType,
		Data:        innerData,
	})
	_ = s.ipc.Write(1, msgData)
}

func (s *Server) Run() {
	for {
		msg, err := s.ipc.Read()
		if err != nil {
			break
		}

		rpcMessage := Msg{}
		_ = json.Unmarshal(msg.Data, &rpcMessage)

		// processing & reply
		process := s.processingCallback[rpcMessage.MessageType]
		if process == nil {
			continue
		}
		replyMsg, err := process(rpcMessage.Data)
		if err != nil {
			replyMsgData, _ := json.Marshal(replyMsg)
			replyMsgData, _ = json.Marshal(Msg{
				Id:          rpcMessage.Id,
				MessageType: -1,
				Data:        []byte(err.Error()),
			})
			_ = s.ipc.Write(2, replyMsgData)
			continue
		}
		if replyMsg == nil {
			continue
		}
		replyMsgData, _ := json.Marshal(replyMsg)
		replyMsgData, _ = json.Marshal(Msg{
			Id:          rpcMessage.Id,
			MessageType: rpcMessage.MessageType,
			Data:        replyMsgData,
		})
		_ = s.ipc.Write(2, replyMsgData)

	}
}

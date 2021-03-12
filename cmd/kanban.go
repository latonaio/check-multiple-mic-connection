package cmd

import (
	"bitbucket.org/latonaio/aion-core/pkg/go-client/msclient"
	"context"
	"log"
	"sync"
)

var kanbanClient msclient.MicroserviceClient
var once sync.Once

const msName = "check-multiple-mic-connection"

func NewKanbanClient(ctx context.Context) error {
	var err error
	once.Do(func() {
		kanbanClient, err = msclient.NewKanbanClient(ctx)
	})
	if err != nil {
		return err
	}

	return nil
}

func KanbanCloseConn() error {
	return kanbanClient.Close()
}

func SetKanban() error {
	_, err := kanbanClient.SetKanban(msName, KanbanProcessNum())
	if err != nil {
		return err
	}
	return nil
}

func KanbanProcessNum() int {
	return kanbanClient.GetProcessNumber()
}

func WriteKanban(data map[string]interface{}, processIndex int, connectionKey string) error {
	metadata := msclient.SetMetadata(data)
	connkey := msclient.SetConnectionKey(connectionKey)
	pNum := msclient.SetProcessNumber(processIndex)
	req, err := msclient.NewOutputData(metadata, pNum, connkey)
	if err != nil {
		return err
	}
	err = kanbanClient.OutputKanban(req)
	if err != nil {
		return err
	}
	log.Println("send msg to kanban")
	return nil
}

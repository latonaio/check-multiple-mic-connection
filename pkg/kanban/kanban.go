package kanban

import (
	"bitbucket.org/latonaio/aion-core/pkg/go-client/msclient"
	"context"
	"log"
	"sync"
)

const msName = "noise-inspection-event-controller"

var (
	client msclient.MicroserviceClient
	once   sync.Once
)

func InitKanbanClient(ctx context.Context) error {
	var err error
	once.Do(func() {
		client, err = msclient.NewKanbanClient(ctx)
		log.Printf("%+v", client)
	})

	return nil
}

func CloseKanban() error {
	return client.Close()
}

func WriteKanban(data map[string]interface{}) error {
	metadata := msclient.SetMetadata(data)
	req, err := msclient.NewOutputData(metadata)
	if err != nil {
		return err
	}
	err = client.OutputKanban(req)
	if err != nil {
		return err
	}
	return nil
}

func GetKanbanCH() (chan *msclient.WrapKanban, error) {
	return client.GetKanbanCh(msName, client.GetProcessNumber())
}

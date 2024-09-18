package userserv

import (
	"fmt"

	"github.com/Alekseyt9/upscaler/internal/back/model"
	"github.com/Alekseyt9/upscaler/internal/back/services/store"
	"github.com/Alekseyt9/upscaler/internal/back/services/websocket"
	cmodel "github.com/Alekseyt9/upscaler/internal/common/model"
	"github.com/Alekseyt9/upscaler/internal/common/services/s3store"
	"golang.org/x/net/context"
)

type UserService struct {
	store   store.Store
	s3store s3store.S3Store
	ws      *websocket.WebSocketService
}

func New(store store.Store, s3store s3store.S3Store, ws *websocket.WebSocketService) *UserService {
	return &UserService{
		store:   store,
		s3store: s3store,
		ws:      ws,
	}
}

func (u *UserService) CreateTasks(ctx context.Context, fileInfos []model.UploadedFile, userID int64) error {
	dlinks, err := u.s3store.GetPresigned(len(fileInfos))
	if err != nil {
		return fmt.Errorf("GetPresigned %w", err)
	}

	tasks := make([]model.StoreTask, 0)
	for i := range fileInfos {
		fileInfo := fileInfos[i]
		dlink := dlinks[i]

		// need to generate URL for downloading
		dURL, err := u.s3store.GetPresignedLoad(fileInfo.Key)
		if err != nil {
			return fmt.Errorf("s3store.GetPresignedLoad %w", err)
		}

		task := model.StoreTask{
			UserID:      userID,
			SrcFileURL:  dURL,
			SrcFileKey:  fileInfo.Key,
			DestFileURL: dlink.Url,
			DestFileKey: dlink.Key,
			FileName:    fileInfo.Name,
		}
		tasks = append(tasks, task)
	}

	err = u.store.CreateTasks(ctx, tasks)
	if err != nil {
		return fmt.Errorf("store.CreateTasks %w", err)
	}

	return nil
}

func (u *UserService) FinishTasks(ctx context.Context, msgs []cmodel.BrokerMessageResult) error {
	var tasks []model.FinishedTask

	for _, m := range msgs {
		url, err := u.s3store.GetPresignedLoad(m.DestFileKey)
		if err != nil {
			return fmt.Errorf("UserService.FinishTasks s3store.GetPresignedLoad %w", err)
		}

		tasks = append(tasks, model.FinishedTask{
			TaskId:      m.TaskId,
			Result:      m.Result,
			Error:       m.Error,
			DestFileURL: url,
		})
	}

	err := u.store.FinishTasks(ctx, tasks)
	if err != nil {
		return fmt.Errorf("store.FinishTasks %w", err)
	}

	for _, m := range msgs {
		err = u.ws.Send(m.UserID, "update")
		if err != nil {
			return fmt.Errorf("userserv FinishTasks ws.Send %w", err)
		}
	}

	return nil
}

package userserv

import (
	"fmt"

	"github.com/Alekseyt9/upscaler/internal/back/model"
	"github.com/Alekseyt9/upscaler/internal/back/services/store"
	cmodel "github.com/Alekseyt9/upscaler/internal/common/model"
	"github.com/Alekseyt9/upscaler/internal/common/services/s3store"
	"golang.org/x/net/context"
)

type UserService struct {
	store   store.Store
	s3store s3store.S3Store
}

func New(store store.Store, s3store s3store.S3Store) *UserService {
	return &UserService{
		store:   store,
		s3store: s3store,
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

		task := model.StoreTask{
			UserID:      userID,
			SrcFileURL:  fileInfo.Url,
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
	err := u.store.FinishTasks(ctx, msgs)
	if err != nil {
		return fmt.Errorf("store.FinishTasks %w", err)
	}
	return nil
}

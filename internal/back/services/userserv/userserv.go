// Package userserv provides the user-related services, including creating file processing tasks,
// completing tasks, and interacting with the S3 store and WebSocket service.
package userserv

import (
	"fmt"
	"strconv"

	"github.com/Alekseyt9/upscaler/internal/back/model"
	"github.com/Alekseyt9/upscaler/internal/back/services/store"
	"github.com/Alekseyt9/upscaler/internal/back/services/websocket"
	cmodel "github.com/Alekseyt9/upscaler/internal/common/model"
	"github.com/Alekseyt9/upscaler/internal/common/services/s3store"
	"golang.org/x/net/context"
)

// UserService defines the interface for user-related services like creating tasks,
// completing tasks, and interacting with the S3 and WebSocket services.
type UserService interface {
	// CreateTasks creates file processing tasks for a user.
	CreateTasks(ctx context.Context, fileInfos []model.UploadedFile, userID int64) error

	// FinishTasks marks file processing tasks as completed and sends updates to users via WebSocket.
	FinishTasks(ctx context.Context, msgs []cmodel.BrokerMessageResult) error
}

// UserService provides user-related services such as creating tasks, finishing tasks,
// and interacting with the WebSocket and S3 storage services.
type UserServiceImpl struct {
	store   store.Store                // Interface to interact with the data store.
	s3store s3store.S3Store            // Interface to interact with S3 storage for file operations.
	ws      websocket.WebSocketService // WebSocket service for real-time communication with users.
}

// New creates and returns a new instance of UserService.
//
// Parameters:
//   - store: Data store to manage tasks and file information.
//   - s3store: S3 storage service to handle file operations.
//   - ws: WebSocket service for notifying users.
//
// Returns:
//   - A pointer to the newly created UserService instance.
func New(store store.Store, s3store s3store.S3Store, ws websocket.WebSocketService) UserService {
	return &UserServiceImpl{
		store:   store,
		s3store: s3store,
		ws:      ws,
	}
}

// CreateTasks creates file processing tasks for a user by generating download and destination URLs
// and storing the tasks in the data store.
//
// Parameters:
//   - ctx: Context for managing request deadlines and cancellation signals.
//   - fileInfos: A slice of uploaded file information.
//   - userID: The ID of the user creating the tasks.
//
// Returns:
//   - An error if there is a failure in generating URLs or creating tasks.
func (u *UserServiceImpl) CreateTasks(ctx context.Context, fileInfos []model.UploadedFile, userID int64) error {
	// Generate presigned URLs for uploading files to S3.
	dlinks, err := u.s3store.GetPresigned(len(fileInfos))
	if err != nil {
		return fmt.Errorf("GetPresigned %w", err)
	}

	tasks := make([]model.StoreTask, 0)
	for i := range fileInfos {
		fileInfo := fileInfos[i]
		dlink := dlinks[i]

		// Generate a download URL for the source file.
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

	_, _, err = u.store.CreateTasks(ctx, tasks)
	if err != nil {
		return fmt.Errorf("store.CreateTasks %w", err)
	}

	return nil
}

// FinishTasks marks file processing tasks as completed and sends real-time updates to users via WebSocket.
//
// Parameters:
//   - ctx: Context for managing request deadlines and cancellation signals.
//   - msgs: A slice of task completion messages from the message broker.
//
// Returns:
//   - An error if there is a failure in updating tasks or sending WebSocket notifications.
func (u *UserServiceImpl) FinishTasks(ctx context.Context, msgs []cmodel.BrokerMessageResult) error {
	var tasks []model.FinishedTask

	for _, m := range msgs {
		// Generate a presigned download URL for the destination file.
		url, err := u.s3store.GetPresignedLoad(m.DestFileKey)
		if err != nil {
			return fmt.Errorf("UserService.FinishTasks s3store.GetPresignedLoad %w", err)
		}

		tasks = append(tasks, model.FinishedTask{
			FileID:      m.FileID,
			Result:      m.Result,
			Error:       m.Error,
			DestFileURL: url,
			UserID:      m.UserID,
			QueueID:     m.QueueID,
		})
	}

	err := u.store.FinishTasks(ctx, tasks)
	if err != nil {
		return fmt.Errorf("store.FinishTasks %w", err)
	}

	for _, m := range msgs {
		err = u.ws.Send(strconv.FormatInt(m.UserID, 10), "update")
		if err != nil {
			return fmt.Errorf("userserv FinishTasks ws.Send %w", err)
		}
	}

	return nil
}

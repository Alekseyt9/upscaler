package cache

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/Alekseyt9/upscaler/internal/back/model"
	"github.com/Alekseyt9/upscaler/internal/back/services/store"
	"github.com/stretchr/testify/require"
)

func TestCachedStore_CreateTasks(t *testing.T) {
	memStore := store.NewMemoryStore()
	var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	cs, err := NewCachedStore(memStore, logger)
	require.NoError(t, err)

	tasks := []model.StoreTask{
		{
			UserID:      1,
			FileName:    "file1.jpg",
			SrcFileURL:  "http://example.com/file1",
			SrcFileKey:  "file1_key",
			DestFileURL: "http://example.com/file1_out",
			DestFileKey: "file1_dest_key",
		},
		{
			UserID:      2,
			FileName:    "file2.jpg",
			SrcFileURL:  "http://example.com/file2",
			SrcFileKey:  "file2_key",
			DestFileURL: "http://example.com/file2_out",
			DestFileKey: "file2_dest_key",
		},
	}

	queueItems, userFileItems, err := cs.CreateTasks(context.Background(), tasks)
	require.NoError(t, err)
	require.Len(t, queueItems, 2)
	require.Len(t, userFileItems, 2)

	userFiles, err := cs.GetUserFiles(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, userFiles, 1)

	userFiles, err = cs.GetUserFiles(context.Background(), 2)
	require.NoError(t, err)
	require.Len(t, userFiles, 1)
}

func TestCachedStore_GetState(t *testing.T) {
	memStore := store.NewMemoryStore()
	var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	cs, err := NewCachedStore(memStore, logger)
	require.NoError(t, err)

	tasks := []model.StoreTask{
		{
			UserID:      1,
			FileName:    "file1.jpg",
			SrcFileURL:  "http://example.com/file1",
			SrcFileKey:  "file1_key",
			DestFileURL: "http://example.com/file1_out",
			DestFileKey: "file1_dest_key",
		},
	}

	_, _, err = cs.CreateTasks(context.Background(), tasks)
	require.NoError(t, err)

	state, err := cs.GetState(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, state, 1)

	require.Equal(t, "file1.jpg", state[0].FileName)
	require.Equal(t, int64(1), state[0].Order)
	require.Equal(t, "PENDING", state[0].Status)
}

func TestCachedStore_FinishTasks(t *testing.T) {
	memStore := store.NewMemoryStore()
	var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	cs, err := NewCachedStore(memStore, logger)
	require.NoError(t, err)

	tasks := []model.StoreTask{
		{
			UserID:      1,
			FileName:    "file1.jpg",
			SrcFileURL:  "http://example.com/file1",
			SrcFileKey:  "file1_key",
			DestFileURL: "http://example.com/file1_out",
			DestFileKey: "file1_dest_key",
		},
	}

	_, fileItems, err := cs.CreateTasks(context.Background(), tasks)
	require.NoError(t, err)
	require.Len(t, fileItems, 1)

	msgs := []model.FinishedTask{
		{
			UserID:      1,
			FileID:      fileItems[0].ID,
			QueueID:     fileItems[0].QueueID,
			DestFileURL: "http://example.com/file1_done",
			Result:      "COMPLETED",
		},
	}

	err = cs.FinishTasks(context.Background(), msgs)
	require.NoError(t, err)

	state, err := cs.GetState(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, "COMPLETED", state[0].Status)
	require.Equal(t, "http://example.com/file1_done", state[0].Link)
}

func TestCachedStore_SendTasksToBroker(t *testing.T) {
	memStore := store.NewMemoryStore()
	var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	cs, err := NewCachedStore(memStore, logger)
	require.NoError(t, err)

	tasks := []model.StoreTask{
		{
			UserID:      1,
			FileName:    "file1.jpg",
			SrcFileURL:  "http://example.com/file1",
			SrcFileKey:  "file1_key",
			DestFileURL: "http://example.com/file1_out",
			DestFileKey: "file1_dest_key",
		},
	}

	_, _, err = cs.CreateTasks(context.Background(), tasks)
	require.NoError(t, err)

	err = cs.SendTasksToBroker(context.Background(), func(items []model.OutboxItem) error {
		require.Len(t, items, 0)
		return nil
	})

	require.NoError(t, err)
}

// Package processor provides the implementation of the ProcessorService,
// which is responsible for handling file processing tasks, including downloading files
// from S3, processing them, and then uploading the results back to S3.
// It also manages task execution through a worker pool and sends the result
// of the processing to a message broker.
package processor

import (
	"context"
	"log/slog"

	"github.com/Alekseyt9/upscaler/internal/common/model"
	"github.com/Alekseyt9/upscaler/internal/common/services/s3store"
	"github.com/Alekseyt9/upscaler/internal/proc/services/fileprocessor"
	"github.com/Alekseyt9/upscaler/internal/proc/services/idcheck"
	"github.com/Alekseyt9/upscaler/internal/proc/services/producer"
	"github.com/Alekseyt9/upscaler/pkg/workerpool"
)

// ProcessorService is responsible for processing files.
type ProcessorService struct {
	wpool    *workerpool.WorkerPool
	s3store  s3store.S3Store
	log      *slog.Logger
	fileproc *fileprocessor.FileProcessor
	idcheck  *idcheck.IdCheckService
	producer *producer.Producer
}

// New creates a new instance of ProcessorService.
//
// Parameters:
//   - wpool: A worker pool to manage task execution.
//   - s3store: An S3Store instance for handling file storage and retrieval.
//   - log: A logger for logging operations and errors.
//   - fileproc: A FileProcessor instance for processing files.
//   - idcheck: An IdCheckService for idempotency checks.
//   - producer: A Producer for sending processing results to Kafka.
//
// Returns:
//   - A pointer to a ProcessorService instance.
func New(wpool *workerpool.WorkerPool, s3store s3store.S3Store,
	log *slog.Logger, fileproc *fileprocessor.FileProcessor,
	idcheck *idcheck.IdCheckService, producer *producer.Producer) *ProcessorService {
	proc := &ProcessorService{
		wpool:    wpool,
		s3store:  s3store,
		log:      log,
		fileproc: fileproc,
		idcheck:  idcheck,
		producer: producer,
	}
	return proc
}

// Process handles the entire process of downloading a file from S3, processing it,
// and uploading the result back to S3. It first checks for idempotency using the idcheck service.
// If the task is valid, it adds the processing task to the worker pool.
//
// Parameters:
//   - ctx: The context for handling request-scoped values and cancellations.
//   - msg: The BrokerMessage containing details of the file to process.
//   - id: The unique identifier for the file, used for idempotency checks.
//
// Returns:
//   - An error if the idempotency check fails or if there is an error during processing.
func (p *ProcessorService) Process(ctx context.Context, msg model.BrokerMessage, id string) error {
	if !p.idcheck.CheckAndSave(ctx, id) {
		return nil
	}

	p.wpool.AddTask(func() {
		path, err := p.s3store.DownloadAndSaveTemp(msg.SrcFileURL, msg.FileExtension)
		if err != nil {
			p.log.Error("s3store.DownloadAndSaveTemp", "error", err)
		}
		p.log.Info("processor Process", "tempfile", path)

		outpath := path + ".out" + msg.FileExtension
		if err == nil {
			err = p.fileproc.Process(path, outpath)
			if err != nil {
				p.log.Error("fileproc.Process", "error", err)
			}
		}

		if err == nil {
			err = p.s3store.Upload(msg.DestFileURL, outpath)
			if err != nil {
				p.log.Error("s3store.Upload", "error", err)
			}
		}

		resMsg := "PROCESSED"
		errMsg := ""
		if err != nil {
			resMsg = "ERROR"
			errMsg = err.Error()
		}

		rmsg := model.BrokerMessageResult{
			FileID:      msg.FileID,
			Result:      resMsg,
			Error:       errMsg,
			DestFileKey: msg.DestFileKey,
			UserID:      msg.UserID,
			QueueID:     msg.QueueID,
		}

		p.log.Info("sended msg", "message", rmsg)

		err = p.producer.Send(rmsg)
		if err != nil {
			p.log.Error("broker.Send", "error", err)
			return
		}

	})

	return nil
}

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

type ProcessorService struct {
	wpool    *workerpool.WorkerPool
	s3store  s3store.S3Store
	log      *slog.Logger
	fileproc *fileprocessor.FileProcessor
	idcheck  *idcheck.IdCheckService
	producer *producer.Producer
}

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

func (p *ProcessorService) Process(ctx context.Context, msg model.BrokerMessage, id string) error {
	if !p.idcheck.CheckAndSave(ctx, id) {
		return nil
	}

	p.wpool.AddTask(func() {
		path, err := p.s3store.DownloadAndSaveTemp(msg.SrcFileURL)
		if err != nil {
			p.log.Error("s3store.DownloadAndSaveTemp", "error", err)
		}
		p.log.Info("processor Process", "tempfile", path)

		outpath := path + ".out"
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

		resMsg := "OK"
		errMsg := ""
		if err != nil {
			resMsg = "ERROR"
			errMsg = err.Error()
		}

		rmsg := model.BrokerMessageResult{
			TaskId: msg.TaskId,
			Result: resMsg,
			Error:  errMsg,
		}
		err = p.producer.Send(rmsg)
		if err != nil {
			p.log.Error("broker.Send", "error", err)
			return
		}

	})

	return nil
}

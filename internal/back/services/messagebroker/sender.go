package messagebroker

import (
	"database/sql"
	"log"
	"time"

	"github.com/IBM/sarama"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Sender struct {
	db       *sql.DB
	producer sarama.SyncProducer
	interval time.Duration
	quit     chan struct{}
}

func NewSender(dbConnStr string, kafkaBrokers []string) (*Sender, error) {
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		return nil, err
	}

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.Flush.Frequency = 500 * time.Millisecond

	producer, err := sarama.NewSyncProducer(kafkaBrokers, config)
	if err != nil {
		return nil, err
	}

	sender := &Sender{
		db:       db,
		producer: producer,
		interval: 3 * time.Second,
		quit:     make(chan struct{}),
	}

	go sender.startSending()

	return sender, nil
}

func (s *Sender) startSending() {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.sendMessagesBatch()
		case <-s.quit:
			log.Println("Остановка отправки сообщений")
			return
		}
	}
}

func (s *Sender) sendMessagesBatch() {
	tx, err := s.db.Begin()
	if err != nil {
		log.Printf("Ошибка при начале транзакции: %v", err)
		return
	}

	rows, err := tx.Query(`
        SELECT id, payload, idempotency_key
        FROM outbox
        WHERE status = 'PENDING'
        ORDER BY created_at
        LIMIT 100
        FOR UPDATE SKIP LOCKED`)
	if err != nil {
		tx.Rollback()
		log.Printf("Ошибка при запросе к outbox: %v", err)
		return
	}
	defer rows.Close()

	var messages []*sarama.ProducerMessage
	var ids []int
	for rows.Next() {
		var id int
		var payload string
		var idempotencyKey string
		if err := rows.Scan(&id, &payload, &idempotencyKey); err != nil {
			tx.Rollback()
			log.Printf("Ошибка при сканировании данных из outbox: %v", err)
			return
		}

		msg := &sarama.ProducerMessage{
			Topic: "file_processing",
			Key:   sarama.StringEncoder(idempotencyKey),
			Value: sarama.StringEncoder(payload),
		}
		messages = append(messages, msg)
		ids = append(ids, id)
	}

	if len(messages) > 0 {
		err = s.producer.SendMessages(messages)
		if err != nil {
			tx.Rollback()
			log.Printf("Ошибка при отправке сообщений: %v", err)
			return
		}

		updateQuery := `
            UPDATE outbox
            SET status = 'SENT', processed_at = NOW()
            WHERE id = ANY($1)`
		_, err = tx.Exec(updateQuery, pq.Array(ids))
		if err != nil {
			tx.Rollback()
			log.Printf("Ошибка при обновлении статуса сообщений: %v", err)
			return
		}

		if err := tx.Commit(); err != nil {
			log.Printf("Ошибка при фиксации транзакции: %v", err)
		} else {
			log.Printf("Отправлено %d сообщений", len(messages))
		}
	} else {
		tx.Rollback()
	}
}

func (s *Sender) Stop() {
	close(s.quit)
	s.db.Close()
	s.producer.Close()
}

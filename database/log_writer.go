package database

import (
	"log"
	"time"
)

const (
	logQueueSize   = 4096
	logBatchSize   = 100
	logFlushPeriod = time.Second
)

var (
	requestLogQueue = make(chan RequestLog, logQueueSize)
	traceLogQueue   = make(chan TraceLog, logQueueSize)
)

// EnqueueRequestLog queues a traffic log for batched insertion.
// It never blocks: when the queue is full the entry is dropped.
func EnqueueRequestLog(l RequestLog) {
	select {
	case requestLogQueue <- l:
	default:
	}
}

// EnqueueTraceLog queues a trace event for batched insertion (non-blocking).
func EnqueueTraceLog(l TraceLog) {
	select {
	case traceLogQueue <- l:
	default:
	}
}

// StartLogWriters starts the background batch writers and the retention purge.
// Entries older than retentionDays are hard-deleted hourly (0 disables purging).
func StartLogWriters(retentionDays int) {
	go batchWrite(requestLogQueue)
	go batchWrite(traceLogQueue)
	if retentionDays > 0 {
		go func() {
			purgeOldLogs(retentionDays)
			for range time.Tick(time.Hour) {
				purgeOldLogs(retentionDays)
			}
		}()
	}
}

func batchWrite[T any](queue <-chan T) {
	batch := make([]T, 0, logBatchSize)
	ticker := time.NewTicker(logFlushPeriod)
	defer ticker.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := GetDB().CreateInBatches(batch, logBatchSize).Error; err != nil {
			log.Printf("Log writer: failed to insert %d entries: %v", len(batch), err)
		}
		batch = batch[:0]
	}

	for {
		select {
		case item := <-queue:
			batch = append(batch, item)
			if len(batch) >= logBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

func purgeOldLogs(retentionDays int) {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	for _, model := range []any{&RequestLog{}, &TraceLog{}} {
		// Unscoped: hard delete, otherwise soft-deleted rows would stay forever.
		res := GetDB().Unscoped().Where("created_at < ?", cutoff).Delete(model)
		if res.Error != nil {
			log.Printf("Log retention: purge failed: %v", res.Error)
		} else if res.RowsAffected > 0 {
			log.Printf("Log retention: purged %d rows older than %d days", res.RowsAffected, retentionDays)
		}
	}
}

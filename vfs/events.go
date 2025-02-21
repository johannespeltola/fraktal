package vfs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Global context for Redis operations.
var ctx = context.Background()

type FileSystemEventType int

const (
	EventCreateFile FileSystemEventType = iota
	EventCreateDir
	EventWriteFile
	EventDelete
)

type FileSystemEvent struct {
	EventType FileSystemEventType `json:"event_type"`
	Path      string              `json:"path"`    // Target path of operation
	Content   string              `json:"content"` // Content for write events
	Timestamp time.Time           `json:"timestamp"`
}

type EventLog struct {
	events []FileSystemEvent
	rdb    *redis.Client
}

// Record event to vfs log and publish to redis
func (log *EventLog) Append(event FileSystemEvent) error {
	log.events = append(log.events, event)

	// No redis client provided - soft exit for easy testing
	if log.rdb == nil {
		return nil
	}

	// Serialize and publish event
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return log.rdb.RPush(ctx, "vfs:events", data).Err()
}

func (log *EventLog) Replay(vfs *VirtualFS) error {
	for _, event := range log.events {
		switch event.EventType {
		case EventCreateFile:
			if err := vfs.CreateFile(event.Path); err != nil {
				return err
			}
		case EventCreateDir:
			if err := vfs.Mkdir(event.Path); err != nil {
				return err
			}
		case EventWriteFile:
			if err := vfs.WriteFile(event.Path, event.Content); err != nil {
				return err
			}
		case EventDelete:
			if err := vfs.Remove(event.Path); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unkown event type: %v", event.EventType)
		}
	}
	return nil
}

// Restore fs events using redis events
func RestoreEventLog(rdb *redis.Client) (*EventLog, error) {
	fmt.Println("--- Restoring file system ---")
	// Get all events from the Redis list.
	data, err := rdb.LRange(context.TODO(), "vfs:events", 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var events []FileSystemEvent
	for _, d := range data {
		var event FileSystemEvent
		if err := json.Unmarshal([]byte(d), &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return &EventLog{events: events, rdb: rdb}, nil
}

func NewEventLog(rdb *redis.Client) *EventLog {
	return &EventLog{
		events: []FileSystemEvent{},
		rdb:    rdb,
	}
}

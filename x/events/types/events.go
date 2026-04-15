package types

import "time"

type Event struct {
    ID        string      `json:"id"`
    Type      string      `json:"type"`
    Data      interface{} `json:"data"`
    Timestamp int64       `json:"timestamp"`
}

func NewEvent(eventType string, data interface{}) Event {
    return Event{
        ID:        "ev-" + time.Now().Format("20060102150405"),
        Type:      eventType,
        Data:      data,
        Timestamp: time.Now().Unix(),
    }
}

package _import

import "time"

type TimestampProvider interface {
	GetTimestamp() string
}

type defaultTimestampProvider struct {
	Format string
}

func DefaultTimestampProvider() TimestampProvider {
	return defaultTimestampProvider{
		Format: time.RFC3339,
	}
}

func (d defaultTimestampProvider) GetTimestamp() string {
	return time.Now().Format(d.Format)
}

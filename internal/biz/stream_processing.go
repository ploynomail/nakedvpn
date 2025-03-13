package biz

import "github.com/go-kratos/kratos/v2/log"

type StreamProcessing struct {
	log *log.Helper
}

func NewStreamProcessing(logger log.Logger) *StreamProcessing {
	return &StreamProcessing{
		log: log.NewHelper(log.With(logger, "module", "biz/stream_processing")),
	}
}

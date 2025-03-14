package biz

import (
	"github.com/go-kratos/kratos/v2/log"
)

type Command uint16

const (
	CommandHeartbeat      Command = iota + 1 // Heartbeat
	CommandReqAuth                           // Request Auth
	CommandInfoCollect                       // Info Collect
	CommandInfoReport                        // Info Report
	CommandAuth                              // Auth
	CommandAuthResult                        // Auth Result
	CommandData                              // Data
	CommandRouteUpdate                       // Route Update
	CommandClose                             // Close
	CommandUpdateSoftware                    // Update Software

	CommandError
)

type HandleUseCase struct {
	log   *log.Helper
	orgUc *OrganizeUseCase
}

func NewHandleUseCase(orgUc *OrganizeUseCase, logger log.Logger) *HandleUseCase {
	return &HandleUseCase{
		log:   log.NewHelper(log.With(logger, "module", "biz/handle")),
		orgUc: orgUc,
	}
}

func (h *HandleUseCase) HandleCommandAuth(orgId uint16, data []byte) (bool, error) {
	if orgId == 0 {
		return false, ErrInvalidOrganize
	}
	if !h.orgUc.AuthAccessKey(orgId, string(data)) {
		return false, nil
	}
	return true, nil
}

func (h *HandleUseCase) HandleCommandData(data []byte) error {
	return nil
}

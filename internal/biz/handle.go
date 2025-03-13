package biz

type Command uint16

const (
	CommandHeartbeat      Command = iota + 1 // Heartbeat
	CommandReqAuth                           // Request Auth
	CommandInfoCollect                       // Info Collect
	CommandInfoReport                        // Info Report
	CommandAuth                              // Auth
	CommandData                              // Data
	CommandRouteUpdate                       // Route Update
	CommandClose                             // Close
	CommandUpdateSoftware                    // Update Software
)

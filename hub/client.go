package hub

import "log/slog"

type (
	Connector interface {
		Alive() bool
		GetSelf() Member
		Listen(func(Message))
		GetGroupMembers(gid string) GroupMembers
		GetGroupMember(gid string, uid string) *GroupMember
		SendText(gid, msg string) (int64, error)
		SendImageByUrl(gid, path, filename string) (int64, error)
		SendImageByBase64(gid string, data []byte, filename string) (int64, error)
		SendVideoByUrl(gid, path, filename string) (int64, error)
		SendVideoByBase64(gid string, data []byte, filename string) (int64, error)
		SendFileByUrl(gid, path, filename string) (int64, error)
		SendFileByBase64(gid string, data []byte, filename string) (int64, error)
	}
)

var (
	Client Connector
	Bot    Member
)

func Init(c Connector) {
	Client = c
	Bot = c.GetSelf()
	slog.Info("Client initialized", "bot", Bot)
}

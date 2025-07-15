package hub

import "encoding/json"

type (
	Message struct {
		BaseMessage
		Content string          `json:"content"`
		Quote   *Quote          `json:"quote,omitempty"`
		At      *At             `json:"at,omitempty"`
		Revoke  *Revoke         `json:"revoke,omitempty"`
		Media   *Media          `json:"media,omitempty"`
		Event   string          `json:"event"`
		Data    json.RawMessage `json:"data"`
	}
	BaseMessage struct {
		MsgType    int    `json:"msgType"`
		Time       int64  `json:"time"`
		MsgID      string `json:"msgID"`
		GID        string `json:"gid,omitempty"`
		GroupName  string `json:"groupName,omitempty"`
		UID        string `json:"uid,omitempty"`
		Username   string `json:"username,omitempty"`
		WechatName string `json:"wechatName,omitempty"`
	}

	Quote struct {
		MsgID   string `json:"msgID"`
		UID     string `json:"uid"`
		Name    string `json:"name"`
		Bot     bool   `json:"bot"`
		Content string `json:"content"`
	}

	At struct {
		UID    string `json:"uid"`
		Name   string `json:"name"`
		Bot    bool   `json:"bot"`
		Offset int    `json:"offset"`
		Length int    `json:"length"`
	}

	// Revoke 撤回信息
	Revoke struct {
		OldMsgID   string `json:"oldMsgID"`
		ReplaceMsg string `json:"replaceMsg"`
	}

	// Media 信息
	Media struct {
		Filename string `json:"filename"`
		Src      string `json:"src"`
		Size     string `json:"size"`
	}

	EventExitGroupUser struct {
		UID  string `json:"uid"`  // 用户id
		Name string `json:"name"` // 用户名
	}
)

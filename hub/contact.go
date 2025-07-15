package hub

type (
	Member struct {
		WxId     string `json:"wxId"`
		NickName string `json:"nickName"`
	}
	GroupMember struct {
		Member
		DisplayName string `json:"displayName"`
	}
	GroupMembers []GroupMember
)

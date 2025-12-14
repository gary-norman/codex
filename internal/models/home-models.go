package models

type HomePage struct {
	UserID                 UUIDField
	CurrentUser            *User
	Instance               string
	Location               string
	UserPosts              []*Post
	AllPosts               []*Post
	OwnedChannels          []*Channel
	JoinedChannels         []*Channel
	OwnedAndJoinedChannels []*Channel
	ThisChannel            *Channel // For edit channel rules popover
	ThisChannelRules       []Rule   // For edit channel rules popover
	ImagePaths
}

func (h HomePage) GetInstance() string { return h.Instance }

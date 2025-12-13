package models

import (
	"time"
)

type Channel struct {
	ID               int64     `db:"id"`
	OwnerID          UUIDField `db:"ownerId"`
	Name             string    `db:"name"`
	Avatar           string    `db:"avatar,omitempty"`
	Banner           string    `db:"banner,omitempty"`
	Description      string    `db:"description"`
	Created          time.Time `db:"created"`
	Updated          time.Time `db:"updated"`
	TimeSince        string
	Rules            []Rule
	UnsubmittedRules []string
	Owned            bool
	Joined           bool
	Privacy          bool `db:"privacy"`
	IsMuted          bool `db:"isMuted"`
	IsFlagged        bool `db:"isFlagged,omitempty"`
	Members          int
	MembersOnline    int
}

func (c Channel) TableName() string { return "channels" }
func (c Channel) GetID() int64      { return c.ID }
func (c *Channel) SetID(id int64)   { c.ID = id }

func (c *Channel) UpdateTimeSince() {
	c.TimeSince = getTimeSince(c.Created)
}

type ChannelPage struct {
	UserID                 UUIDField
	CurrentUser            *User
	Instance               string
	Location               string
	ThisChannel            *Channel
	OwnerName              string
	IsOwned                bool
	IsJoined               bool
	Rules                  []Rule
	Posts                  []*Post
	OwnedChannels          []*Channel
	JoinedChannels         []*Channel
	OwnedAndJoinedChannels []*Channel
	ImagePaths
}

func (p ChannelPage) GetInstance() string { return p.Instance }

type ChannelData struct {
	ChannelName string `db:"channelName"`
	ChannelID   string `db:"channelId"`
}

func (cd ChannelData) TableName() string { return "channel_data" }

type MutedChannel struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"userId"`
	ChannelID int64     `db:"channelId"`
	Created   time.Time `db:"created"`
}

func (m MutedChannel) TableName() string { return "muted_channels" }
func (m MutedChannel) GetID() int64      { return m.ID }
func (m *MutedChannel) SetID(id int64)   { m.ID = id }

type OwnedAndJoinedChannels struct {
	Owned    bool
	Joined   bool
	Channels []Channel
}

type Membership struct {
	ID        int64     `db:"id"`
	UserID    UUIDField `db:"userId"`
	ChannelID int64     `db:"channelId"`
	Created   time.Time `db:"created"`
	Updated   time.Time `db:"updated"`
}

func (m Membership) TableName() string { return "memberships" }
func (m Membership) GetID() int64      { return m.ID }
func (m *Membership) SetID(id int64)   { m.ID = id }

type Mod struct {
	ID        int64     `db:"id"`
	UserID    UUIDField `db:"userId"`
	ChannelID int64     `db:"channelId"`
	Created   time.Time `db:"created"`
	TimeSince string
}

func (m Mod) TableName() string { return "mods" }
func (m Mod) GetID() int64      { return m.ID }
func (m *Mod) SetID(id int64)   { m.ID = id }

func (m *Mod) UpdateTimeSince() {
	m.TimeSince = getTimeSince(m.Created)
}

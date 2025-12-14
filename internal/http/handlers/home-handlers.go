// Package handlers contains the HTTP handlers for the forum application.
package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gary-norman/forum/internal/app"
	// "github.com/gary-norman/forum/internal/dao"
	mw "github.com/gary-norman/forum/internal/http/middleware"
	"github.com/gary-norman/forum/internal/models"
	"github.com/gary-norman/forum/internal/view"
)

type HomeHandler struct {
	App      *app.App
	Reaction *ReactionHandler
	Post     *PostHandler
	Comment  *CommentHandler
	Channel  *ChannelHandler
	Mod      *ModHandler
}

var TemplateData models.TemplateData

func (h *HomeHandler) RenderIndex(w http.ResponseWriter, r *http.Request) {
	var ErrorPage bool

	start := time.Now()
	fmt.Println(r.PathValue("invalidString"))
	stringCheck := []string{" ", "favicon.ico"}
	if r.PathValue("invalidString") != "" {
		valid := false
		for _, v := range stringCheck {
			fmt.Println("checking against valid string:", v)
			if r.PathValue("invalidString") == v {
				valid = true
				break
			}
		}
		if !valid {
			fmt.Printf("illegal string '%v' detected\n", r.PathValue("invalidString"))
			ErrorPage = true
			w.WriteHeader(400)
			// ErrorPage = models.ErrorPage{
			// 	Data:   models.NotFoundLocation("home"),
			// 	Status: 400,
			// 	Error:  models.NotFoundError(r.PathValue("invalidString"), "RenderIndex", fmt.Errorf("the user entered %v after /", r.PathValue("invalidString"))),
			// }
		}
	}

	userLoggedIn := true
	userID := models.ZeroUUIDField()
	// SECTION --- posts and comments ---
	// postDAO := dao.NewDAO[*models.Post](h.App.DB)
	// ctx := context.Background()
	// daoAllPosts, err := postDAO.All(ctx)
	// if err != nil {
	// 	log.Printf(ErrorMsgs.KeyValuePair, "Error fetching daoAllposts", err)
	// } else {
	// 	for _, post := range daoAllPosts {
	// 		fmt.Printf(ErrorMsgs.KeyValuePair, "daoAllPosts", post.Title)
	// 	}
	// }

	// SECTION --- user ---
	allUsers, allUsersErr := h.App.Users.All()
	if allUsersErr != nil {
		log.Printf(ErrorMsgs.Query, "RenderIndex> users > All", allUsersErr)
	}
	for u := range allUsers {
		models.UpdateTimeSince(allUsers[u])
	}

	// attach following/follower numbers to each user
	for u := range allUsers {
		allUsers[u].Followers, allUsers[u].Following, allUsersErr = h.App.Loyalty.CountUsers(allUsers[u].ID)
		if allUsersErr != nil {
			log.Printf(ErrorMsgs.Query, "RenderIndex> users > All > loyalty", allUsersErr)
		}
	}

	randomUser := GetRandomUser(allUsers)
	currentUser, ok := mw.GetUserFromContext(r.Context())
	if !ok {
		userLoggedIn = false
	}

	var currentUserErr error
	// attach following/follower numbers to the random user
	randomUser.Followers, randomUser.Following, currentUserErr = h.App.Loyalty.CountUsers(randomUser.ID)
	if currentUserErr != nil {
		log.Printf(ErrorMsgs.Query, "RenderIndex> users > All", allUsersErr)
	}

	// SECTION --- posts and comments ---
	allPosts, err := h.App.Posts.All()
	if err != nil {
		log.Printf(ErrorMsgs.KeyValuePair, "Error fetching all posts", err)
	}
	// Retrieve total likes and dislikes for each post
	allPosts = h.Reaction.GetPostsLikesAndDislikes(allPosts)

	// Retrieve last reaction time for posts
	allPosts, err = h.Reaction.getLastReactionTimeForPosts(allPosts)
	if err != nil {
		log.Printf(ErrorMsgs.KeyValuePair, "Error getting last reaction time for allPosts", err)
	}

	for p := range allPosts {
		models.UpdateTimeSince(allPosts[p])
	}
	allPosts, err = h.Comment.GetPostsComments(allPosts)
	if err != nil {
		log.Printf(ErrorMsgs.NotFound, "allPosts comments", "RenderIndex", err)
	}

	for p := range allPosts {
		channelIDs, err := h.App.Channels.GetChannelIDFromPost(allPosts[p].ID)
		if err != nil {
			log.Printf(ErrorMsgs.KeyValuePair, "RenderIndex > channelID", err)
		}
		if len(allPosts) > 0 && len(channelIDs) > 0 {
			allPosts[p].ChannelID = channelIDs[0]
		} else {
			fetchErr := fmt.Sprintf("post ID: %v does not belong to any channel", allPosts[p].ID)
			fmt.Printf(ErrorMsgs.KeyValuePair, "error fetching posts", fetchErr)
		}
	}

	// SECTION --- channels --
	allChannels, err := h.App.Channels.All()
	if err != nil {
		log.Printf(ErrorMsgs.Query, "channels.All-index", err)
	}
	for c := range allChannels {
		models.UpdateTimeSince(allChannels[c])
	}

	for p := range allPosts {
		for _, channel := range allChannels {
			if channel.ID == allPosts[p].ChannelID {
				allPosts[p].ChannelName = channel.Name
			}
		}
	}

	ownedChannels := make([]*models.Channel, 0)
	joinedChannels := make([]*models.Channel, 0)
	ownedAndJoinedChannels := make([]*models.Channel, 0)
	channelMap := make(map[int64]bool)

	if userLoggedIn {
		userID = currentUser.ID
		// attach following/follower numbers to currently logged-in user
		currentUser.Followers, currentUser.Following, err = h.App.Loyalty.CountUsers(currentUser.ID)
		if err != nil {
			fmt.Printf(ErrorMsgs.KeyValuePair, "RenderIndex > currentUser loyalty", err)
		}
		// get owned and joined channels of current user
		memberships, memberErr := h.App.Memberships.UserMemberships(currentUser.ID)
		if memberErr != nil {
			log.Printf(ErrorMsgs.KeyValuePair, "RenderIndex > UserMemberships", memberErr)
		}
		ownedChannels, err = h.App.Channels.OwnedOrJoinedByCurrentUser(currentUser.ID)
		if err != nil {
			log.Printf(ErrorMsgs.Query, "RenderIndex > user owned channels", err)
		}
		joinedChannels, err = h.Channel.JoinedByCurrentUser(memberships)
		if err != nil {
			log.Printf(ErrorMsgs.Query, "user joined channels", err)
		}

		// ownedAndJoinedChannels = append(ownedChannels, joinedChannels...)
		// Add owned channels
		for _, channel := range ownedChannels {
			if !channelMap[channel.ID] {
				channelMap[channel.ID] = true
				ownedAndJoinedChannels = append(ownedAndJoinedChannels, channel)
			}
		}

		// Add joined channels
		for _, channel := range joinedChannels {
			if !channelMap[channel.ID] {
				channelMap[channel.ID] = true
				ownedAndJoinedChannels = append(ownedAndJoinedChannels, channel)
			}
		}
	} else {
		ownedAndJoinedChannels = allChannels
	}

	// SECTION -- chats ---
	var chats []models.Chat

	if userLoggedIn {
		// GetUserChats for the current user
		chats, err = h.App.Chats.GetUserChats(currentUser.ID)
		if err != nil {
			log.Printf(ErrorMsgs.Query, "RenderIndex > GetUserChats", err)
		}

		// GetChatMessages for each chat
		for _, chat := range chats {
			messages, err := h.App.Chats.GetChatMessages(chat.ID)
			if err != nil {
				log.Printf(ErrorMsgs.Query, "RenderIndex > GetChatMessages", err)
			}
			chat.Messages = messages
		}
	}

	fakeChatID_1 := models.NewUUIDField()
	fakeChatID_2 := models.NewUUIDField()
	fakeUser_1 := allUsers[1]
	fakeUser_2 := allUsers[2]
	fakeUser_3 := allUsers[3]
	fakeUser_4 := allUsers[4]

	chats = append(chats, models.Chat{
		ID:         fakeChatID_1,
		ChatType:   "buddy",
		Name:       "chat logic",
		LastActive: time.Now(),
		Buddy:      fakeUser_1,
		Messages: []models.ChatMessage{
			{ChatID: fakeChatID_1, Sender: currentUser, Content: "Hey there! What's up", Created: time.Now().Add(-24*time.Hour - 90*time.Second)},
			{ChatID: fakeChatID_1, Sender: fakeUser_1, Content: "Checking out this new chat..", Created: time.Now().Add(-24*time.Hour - 75*time.Second)},
			{ChatID: fakeChatID_1, Sender: currentUser, Content: "Check out this bubble!", Created: time.Now().Add(-24*time.Hour - 60*time.Second)},
			{ChatID: fakeChatID_1, Sender: fakeUser_1, Content: "It's pretty cool…", Created: time.Now().Add(-24*time.Hour - 45*time.Second)},
			{ChatID: fakeChatID_1, Sender: fakeUser_1, Content: "Not gonna lie!", Created: time.Now().Add(-24*time.Hour - 30*time.Second)},
			{ChatID: fakeChatID_1, Sender: currentUser, Content: "Yeah it's pure CSS & HTML", Created: time.Now().Add(-24*time.Hour - 15*time.Second)},
			{ChatID: fakeChatID_1, Sender: fakeUser_1, Content: "Wow that's impressive. But what's even more impressive is that this bubble is really high.", Created: time.Now().Add(-24 * time.Hour)},
			{ChatID: fakeChatID_1, Sender: currentUser, Content: "popover id`\"form-chat-{{ $chat.ID }}\"`", Created: time.Now().Add(-105 * time.Second)},
			{ChatID: fakeChatID_1, Sender: fakeUser_1, Content: "You mean the popovers are dynamically created?", Created: time.Now().Add(-90 * time.Second)},
			{ChatID: fakeChatID_1, Sender: currentUser, Content: "Yes! As are the buttons in the sidebar", Created: time.Now().Add(-75 * time.Second)},
			{ChatID: fakeChatID_1, Sender: fakeUser_1, Content: "So are these chats stored in the database?", Created: time.Now().Add(-60 * time.Second)},
			{ChatID: fakeChatID_1, Sender: currentUser, Content: "They sure are. Check out chats-sql.go", Created: time.Now().Add(-45 * time.Second)},
			{ChatID: fakeChatID_1, Sender: fakeUser_1, Content: "Great! We should have this working pretty soon then!", Created: time.Now().Add(-30 * time.Second)},
			{ChatID: fakeChatID_1, Sender: currentUser, Content: "I think so yes!", Created: time.Now().Add(-15 * time.Second)},
		},
	})

	chats = append(chats, models.Chat{
		ID:         fakeChatID_2,
		ChatType:   "group",
		Name:       "languages",
		LastActive: time.Now(),
		Group:      models.Group{ID: models.NewUUIDField(), Name: "Language Lovers"},
		Messages: []models.ChatMessage{
			{ChatID: fakeChatID_2, Sender: fakeUser_1, Content: "JavaScript is obviously the best language. It runs everywhere!", Created: time.Now().Add(-24*time.Hour - 210*time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_2, Content: "Come on, TypeScript is just JavaScript but actually good. Type safety matters!", Created: time.Now().Add(-24*time.Hour - 195*time.Second)},
			{ChatID: fakeChatID_2, Sender: currentUser, Content: "You both are missing the point. Go is simple, fast, and has amazing concurrency primitives.", Created: time.Now().Add(-24*time.Hour - 180*time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_3, Content: "Rust enters the chat. Memory safety without garbage collection? That's the future.", Created: time.Now().Add(-24*time.Hour - 165*time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_4, Content: "You're all using training wheels. Real programmers write Assembly.", Created: time.Now().Add(-24*time.Hour - 150*time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_1, Content: "Assembly? What is this, 1985? JavaScript has npm with millions of packages!", Created: time.Now().Add(-24*time.Hour - 135*time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_2, Content: "Yeah and half of them are broken or abandoned. At least TypeScript catches errors at compile time.", Created: time.Now().Add(-24*time.Hour - 120*time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_3, Content: "Cargo is way better than npm. And the borrow checker prevents entire classes of bugs.", Created: time.Now().Add(-24*time.Hour - 105*time.Second)},
			{ChatID: fakeChatID_2, Sender: currentUser, Content: "Go's tooling is unmatched though. go fmt, go test, everything just works out of the box.", Created: time.Now().Add(-24*time.Hour - 90*time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_4, Content: "You know what just works? MOV instructions. No abstractions, just pure speed.", Created: time.Now().Add(-24*time.Hour - 75*time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_1, Content: "Speed doesn't matter if development takes forever. I can prototype in JS in minutes.", Created: time.Now().Add(-24*time.Hour - 60*time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_2, Content: "And then spend weeks debugging runtime errors that TypeScript would have caught.", Created: time.Now().Add(-24*time.Hour - 45*time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_3, Content: "If it compiles in Rust, it usually works. Can't say that about any of your languages.", Created: time.Now().Add(-24*time.Hour - 30*time.Second)},
			{ChatID: fakeChatID_2, Sender: currentUser, Content: "Rust's learning curve is brutal though. Go is productive from day one.", Created: time.Now().Add(-24*time.Hour - 15*time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_4, Content: "You're all arguing about high-level languages while I'm optimizing cache lines.", Created: time.Now().Add(-24 * time.Hour)},
			{ChatID: fakeChatID_2, Sender: fakeUser_1, Content: "Okay but seriously, can any of your languages run in a browser natively?", Created: time.Now().Add(-195 * time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_3, Content: "WebAssembly exists. Rust compiles to it beautifully.", Created: time.Now().Add(-180 * time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_2, Content: "And TypeScript compiles to JavaScript, so yes, it runs in browsers too.", Created: time.Now().Add(-165 * time.Second)},
			{ChatID: fakeChatID_2, Sender: currentUser, Content: "Go can compile to WASM too. But honestly, use the right tool for the job.", Created: time.Now().Add(-150 * time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_4, Content: "The right tool is always Assembly. Fight me.", Created: time.Now().Add(-135 * time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_1, Content: "Nobody is fighting you, we're just ignoring you 😂", Created: time.Now().Add(-120 * time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_3, Content: "Can we all agree that at least we're not using PHP?", Created: time.Now().Add(-105 * time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_2, Content: "Now THAT we can agree on!", Created: time.Now().Add(-90 * time.Second)},
			{ChatID: fakeChatID_2, Sender: currentUser, Content: "Lol fair point. Though modern PHP isn't terrible...", Created: time.Now().Add(-75 * time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_1, Content: "Don't defend PHP! You're a Go developer!", Created: time.Now().Add(-60 * time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_4, Content: "PHP is written in C. C compiles to Assembly. Therefore PHP is Assembly.", Created: time.Now().Add(-45 * time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_3, Content: "That logic is so broken even Rust's borrow checker couldn't save it.", Created: time.Now().Add(-30 * time.Second)},
			{ChatID: fakeChatID_2, Sender: fakeUser_2, Content: "I think we can all agree this conversation needs better type definitions.", Created: time.Now().Add(-15 * time.Second)},
		},
	})

	// SECTION -- template ---
	data := models.TemplateData{
		// ---------- users ----------
		UserID:      userID,
		AllUsers:    allUsers,
		RandomUser:  randomUser,
		CurrentUser: currentUser,
		// ---------- channels ----------
		AllChannels:            allChannels,
		OwnedChannels:          ownedChannels,
		JoinedChannels:         joinedChannels,
		OwnedAndJoinedChannels: ownedAndJoinedChannels,
		// ---------- chats ----------
		Chats: chats,
		// ---------- misc ----------
		Instance:   "home-page",
		ImagePaths: h.App.Paths,
		ErrorPage:  ErrorPage,
	}
	log.Printf(ErrorMsgs.KeyValuePair, "RenderIndex > data.UserID", data.UserID)
	// models.JsonError(TemplateData)
	tpl, err := view.GetTemplate()
	if err != nil {
		log.Printf(ErrorMsgs.Parse, "templates", "RenderIndex", err)
		return
	}

	t := tpl.Lookup("index.html")

	if t == nil {
		http.Error(w, "Template is not initialized", http.StatusInternalServerError)
		return
	}

	execErr := t.Execute(w, data)
	if execErr != nil {
		log.Printf(ErrorMsgs.Execute, execErr)
		return
	}
	log.Printf(ErrorMsgs.KeyValuePair, "RenderIndex Render time", time.Since(start))
}

func (h *HomeHandler) GetHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	start := time.Now()
	var userPosts []*models.Post
	userLoggedIn := true
	userID := models.ZeroUUIDField()

	// SECTION --- user ---
	allUsers, allUsersErr := h.App.Users.All()
	if allUsersErr != nil {
		log.Printf(ErrorMsgs.Query, "getHome> users > All", allUsersErr)
	}
	for u := range allUsers {
		models.UpdateTimeSince(allUsers[u])
	}

	// attach following/follower numbers to each user
	for u := range allUsers {
		allUsers[u].Followers, allUsers[u].Following, allUsersErr = h.App.Loyalty.CountUsers(allUsers[u].ID)
		if allUsersErr != nil {
			log.Printf(ErrorMsgs.Query, "getHome> users > All > loyalty", allUsersErr)
		}
	}

	randomUser := GetRandomUser(allUsers)
	currentUser, ok := mw.GetUserFromContext(r.Context())
	if !ok {
		userLoggedIn = false
	}

	var currentUserErr error
	// attach following/follower numbers to the random user
	randomUser.Followers, randomUser.Following, currentUserErr = h.App.Loyalty.CountUsers(randomUser.ID)
	if currentUserErr != nil {
		log.Printf(ErrorMsgs.Query, "getHome> users > All", allUsersErr)
	}

	// SECTION --- posts and comments ---
	allPosts, err := h.App.Posts.All()
	if err != nil {
		log.Printf(ErrorMsgs.KeyValuePair, "Error fetching all posts", err)
	}
	// Retrieve total likes and dislikes for each post
	allPosts = h.Reaction.GetPostsLikesAndDislikes(allPosts)

	// Retrieve last reaction time for posts
	allPosts, err = h.Reaction.getLastReactionTimeForPosts(allPosts)
	if err != nil {
		log.Printf(ErrorMsgs.KeyValuePair, "Error getting last reaction time for allPosts", err)
	}

	for p := range allPosts {
		models.UpdateTimeSince(allPosts[p])
	}
	allPosts, err = h.Comment.GetPostsComments(allPosts)
	if err != nil {
		log.Printf(ErrorMsgs.NotFound, "allPosts comments", "getHome", err)
	}

	for p := range allPosts {
		channelIDs, err := h.App.Channels.GetChannelIDFromPost(allPosts[p].ID)
		if err != nil {
			log.Printf(ErrorMsgs.KeyValuePair, "getHome > channelID", err)
		}
		if len(allPosts) > 0 && len(channelIDs) > 0 {
			allPosts[p].ChannelID = channelIDs[0]
		} else {
			fetchErr := fmt.Sprintf("post ID: %v does not belong to any channel", allPosts[p].ID)
			fmt.Printf(ErrorMsgs.KeyValuePair, "error fetching posts", fetchErr)
		}
	}

	// SECTION --- channels --
	allChannels, err := h.App.Channels.All()
	if err != nil {
		log.Printf(ErrorMsgs.Query, "channels.All-home", err)
	}
	for c := range allChannels {
		models.UpdateTimeSince(allChannels[c])
	}

	for p := range allPosts {
		for _, channel := range allChannels {
			if channel.ID == allPosts[p].ChannelID {
				allPosts[p].ChannelName = channel.Name
			}
		}
	}

	ownedChannels := make([]*models.Channel, 0)
	joinedChannels := make([]*models.Channel, 0)
	ownedAndJoinedChannels := make([]*models.Channel, 0)
	channelMap := make(map[int64]bool)

	if userLoggedIn {
		userID = currentUser.ID
		userPosts = h.Post.GetUserPosts(currentUser, allPosts)
		// attach following/follower numbers to currently logged-in user
		currentUser.Followers, currentUser.Following, err = h.App.Loyalty.CountUsers(currentUser.ID)
		if err != nil {
			fmt.Printf(ErrorMsgs.KeyValuePair, "getHome > currentUser loyalty", err)
		}
		// get owned and joined channels of current user
		memberships, memberErr := h.App.Memberships.UserMemberships(currentUser.ID)
		if memberErr != nil {
			log.Printf(ErrorMsgs.KeyValuePair, "getHome > UserMemberships", memberErr)
		}
		ownedChannels, err = h.App.Channels.OwnedOrJoinedByCurrentUser(currentUser.ID)
		if err != nil {
			log.Printf(ErrorMsgs.Query, "GetHome > user owned channels", err)
		}
		joinedChannels, err = h.Channel.JoinedByCurrentUser(memberships)
		if err != nil {
			log.Printf(ErrorMsgs.Query, "user joined channels", err)
		}

		// ownedAndJoinedChannels = append(ownedChannels, joinedChannels...)
		// Add owned channels
		for _, channel := range ownedChannels {
			if !channelMap[channel.ID] {
				channelMap[channel.ID] = true
				ownedAndJoinedChannels = append(ownedAndJoinedChannels, channel)
			}
		}

		// Add joined channels
		for _, channel := range joinedChannels {
			if !channelMap[channel.ID] {
				channelMap[channel.ID] = true
				ownedAndJoinedChannels = append(ownedAndJoinedChannels, channel)
			}
		}

	} else {
		userPosts = allPosts
		ownedAndJoinedChannels = allChannels
	}

	// SECTION -- template ---
	data := models.HomePage{
		// ---------- users ----------
		UserID:      userID,
		CurrentUser: currentUser,
		// ---------- posts ----------
		AllPosts:  allPosts,
		UserPosts: userPosts,
		// ---------- channels ----------
		OwnedChannels:          ownedChannels,
		JoinedChannels:         joinedChannels,
		OwnedAndJoinedChannels: ownedAndJoinedChannels,
		// ---------- misc ----------
		Instance:   "home-page",
		ImagePaths: h.App.Paths,
	}

	view.RenderPageData(w, data)
	log.Printf(ErrorMsgs.KeyValuePair, "GetHome Render time", time.Since(start))
}

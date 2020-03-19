package simulator

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/matt035343/devops/app/src/middleware"
	"github.com/matt035343/devops/app/src/types"
	"github.com/matt035343/devops/app/src/utils"
)

func (c *Controller) getUserIDFromUrl(r *http.Request) (int, error) {
	username := mux.Vars(r)["username"]
	u, err := c.DB.GetUserFromUsername(username)
	return u.UserID, err
}

func (c *Controller) getUserID(username string) (int, error) {
	u, err := c.DB.GetUserFromUsername(username)
	if err != nil {
		return 0, err
	}
	return u.UserID, nil
}

func (c *Controller) tweetsGet(w http.ResponseWriter, r *http.Request) {
	latest, latestErr := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)
	noMsgs, noMsgsErr := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)
	if latestErr != nil || noMsgsErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		c.DB.SetLatest(latest)
		messages, err := c.DB.GetPublicViewMessages(int(noMsgs))
		if err != nil {
			panic(err)
		}

		filteredMsgs := types.ConvertToTweetResponse(messages)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(filteredMsgs)
	}
}

func (c *Controller) tweetsUsername(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.tweetsUsernameGet(w, r)
	} else if r.Method == "POST" {
		c.tweetsUsernamePost(w, r)
	}
}

func (c *Controller) tweetsUsernameGet(w http.ResponseWriter, r *http.Request) {
	latest, latestErr := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)
	noMsgs, noMsgsErr := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)

	if latestErr != nil || noMsgsErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID, userIDErr := c.getUserIDFromUrl(r)
	if userIDErr != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {

		c.DB.SetLatest(latest)
		messages, err := c.DB.GetUserViewMessages(userID, int(noMsgs))
		if err != nil {
			panic(err)
		}

		filteredMsgs := types.ConvertToTweetResponse(messages)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(filteredMsgs)
	}
}

func (c *Controller) tweetsUsernamePost(w http.ResponseWriter, r *http.Request) {
	userID, userIDErr := c.getUserIDFromUrl(r)
	if userIDErr != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var tweet types.TweetRequest
	tweetErr := decoder.Decode(&tweet)
	if tweetErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		err := c.DB.AddMessage(userID, tweet.Content, time.Now())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		middleware.MessagesSent.Inc()
		w.WriteHeader(http.StatusNoContent)
	}
}

func (c *Controller) followUsername(w http.ResponseWriter, r *http.Request) {
	latest, latestErr := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)

	if latestErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	c.DB.SetLatest(latest)
	userID, userIDErr := c.getUserIDFromUrl(r)
	if userIDErr != nil || userID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method == "GET" {
		c.followUsernameGet(w, r, userID)
	} else if r.Method == "POST" {
		decoder := json.NewDecoder(r.Body)
		var followRequest types.FollowRequest
		decoder.Decode(&followRequest)
		if followRequest.Follow == "" && followRequest.Unfollow == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		isFollow := followRequest.Follow != ""
		if isFollow {
			c.followUsernamePost(w, r, userID, followRequest)
		} else {
			c.unFollowUsernamePost(w, r, userID, followRequest)
		}
	}
}

func (c *Controller) followUsernameGet(w http.ResponseWriter, r *http.Request, userID int) {
	noFollowers, noMsgsErr := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)
	if noMsgsErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	followers, err := c.DB.GetFollowers(userID, int(noFollowers))
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(followers)
}

func (c *Controller) followUsernamePost(w http.ResponseWriter, r *http.Request, userID int, followRequest types.FollowRequest) {
	followsUserID, followsUserErr := c.getUserID(followRequest.Follow)
	if followsUserErr != nil || followsUserID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		followInsertErr := c.DB.AddFollower(userID, followsUserID)
		if followInsertErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		middleware.UsersFollowed.Inc()
		w.WriteHeader(http.StatusNoContent)
	}
}

func (c *Controller) unFollowUsernamePost(w http.ResponseWriter, r *http.Request, userID int, unfollowRequest types.FollowRequest) {
	unfollowsUserID, unfollowsUserErr := c.getUserID(unfollowRequest.Unfollow)
	if unfollowsUserErr != nil || unfollowsUserID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		unfollowErr := c.DB.DeleteFollower(userID, unfollowsUserID)
		if unfollowErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		middleware.UsersUnfollowed.Inc()
		w.WriteHeader(http.StatusNoContent)
	}
}

func (c *Controller) latest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	l, err := c.DB.GetLatest()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(l)
}

func (c *Controller) register(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var registerRequest types.RegisterRequest
	decoder.Decode(&registerRequest)
	if registerRequest == (types.RegisterRequest{}) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	latest, latestErr := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 64)
	if latestErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	c.DB.SetLatest(latest)
	error := ""
	if registerRequest.Username == "" {
		error = utils.EnterAUsername
	} else if registerRequest.Email == "" || !strings.Contains(registerRequest.Email, "@") {
		error = utils.EnterAValidEmail
	} else if registerRequest.Pwd == "" {
		error = utils.YouHaveToEnterAPassword
	} else if !c.DB.IsUsernameAvailable(registerRequest.Username) {
		error = utils.UsernameTaken
	}
	if error != "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.ErrorMsgResponse{Status: 400, ErrorMsg: error})
		return
	} else {
		res := c.registerUser(registerRequest.Username, registerRequest.Email, registerRequest.Pwd)
		if res {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func (c *Controller) registerUser(username string, email string, hashedPassword string) bool {
	err := c.DB.AddUser(username, email, hashedPassword)
	if err != nil {
		return false
	}
	middleware.UsersRegistered.Inc()
	return true
}

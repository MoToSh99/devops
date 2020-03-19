package simulator

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/matt035343/devops/app/src/middleware"
	"github.com/matt035343/devops/app/src/types"
)

func (s *SimulatorServer) getUserIDFromUrl(r *http.Request) (int, error) {
	username := mux.Vars(r)["username"]
	u, err := s.DB.GetUserFromUsername(username)
	return u.UserID, err
}

func (s *SimulatorServer) getUserID(username string) (int, error) {
	u, err := s.DB.GetUserFromUsername(username)
	if err != nil {
		return 0, err
	}
	return u.UserID, nil
}

func (s *SimulatorServer) tweetsGet(w http.ResponseWriter, r *http.Request) {
	latest, latestErr := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)
	noMsgs, noMsgsErr := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)
	if latestErr != nil || noMsgsErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		s.DB.SetLatest(latest)
		messages, err := s.DB.GetPublicViewMessages(int(noMsgs))
		if err != nil {
			panic(err)
		}

		filteredMsgs := types.ConvertToTweetResponse(messages)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(filteredMsgs)
	}
}

func (s *SimulatorServer) tweetsUsername(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		s.tweetsUsernameGet(w, r)
	} else if r.Method == "POST" {
		s.tweetsUsernamePost(w, r)
	}
}

func (s *SimulatorServer) tweetsUsernameGet(w http.ResponseWriter, r *http.Request) {
	latest, latestErr := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)
	noMsgs, noMsgsErr := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)

	if latestErr != nil || noMsgsErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID, userIDErr := s.getUserIDFromUrl(r)
	if userIDErr != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {

		s.DB.SetLatest(latest)
		messages, err := s.DB.GetUserViewMessages(userID, int(noMsgs))
		if err != nil {
			panic(err)
		}

		filteredMsgs := types.ConvertToTweetResponse(messages)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(filteredMsgs)
	}
}

func (s *SimulatorServer) tweetsUsernamePost(w http.ResponseWriter, r *http.Request) {
	userID, userIDErr := s.getUserIDFromUrl(r)
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
		err := s.DB.AddMessage(userID, tweet.Content, time.Now())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		middleware.MessagesSent.Inc()
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *SimulatorServer) followUsername(w http.ResponseWriter, r *http.Request) {
	latest, latestErr := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)

	if latestErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s.DB.SetLatest(latest)
	userID, userIDErr := s.getUserIDFromUrl(r)
	if userIDErr != nil || userID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method == "GET" {
		s.followUsernameGet(w, r, userID)
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
			s.followUsernamePost(w, r, userID, followRequest)
		} else {
			s.unFollowUsernamePost(w, r, userID, followRequest)
		}
	}
}

func (s *SimulatorServer) followUsernameGet(w http.ResponseWriter, r *http.Request, userID int) {
	noFollowers, noMsgsErr := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)
	if noMsgsErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	followers, err := s.DB.GetFollowers(userID, int(noFollowers))
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(followers)
}

func (s *SimulatorServer) followUsernamePost(w http.ResponseWriter, r *http.Request, userID int, followRequest types.FollowRequest) {
	followsUserID, followsUserErr := s.getUserID(followRequest.Follow)
	if followsUserErr != nil || followsUserID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		followInsertErr := s.DB.AddFollower(userID, followsUserID)
		if followInsertErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		middleware.UsersFollowed.Inc()
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *SimulatorServer) unFollowUsernamePost(w http.ResponseWriter, r *http.Request, userID int, unfollowRequest types.FollowRequest) {
	unfollowsUserID, unfollowsUserErr := s.getUserID(unfollowRequest.Unfollow)
	if unfollowsUserErr != nil || unfollowsUserID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		unfollowErr := s.DB.DeleteFollower(userID, unfollowsUserID)
		if unfollowErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		middleware.UsersUnfollowed.Inc()
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *SimulatorServer) latest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	l, err := s.DB.GetLatest()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(l)
}

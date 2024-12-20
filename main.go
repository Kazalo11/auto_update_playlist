package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var (
	redirectURI = os.Getenv("SPOTIFY_REDIRECT_URL")
	auth        = spotifyauth.New(
		spotifyauth.WithClientID(os.Getenv("SPOTIFY_ID")),
		spotifyauth.WithClientSecret(os.Getenv("SPOTIFY_SECRET")),
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(spotifyauth.ScopeUserLibraryRead, spotifyauth.ScopePlaylistModifyPrivate, spotifyauth.ScopePlaylistModifyPublic),
	)
	state = "abc123"
	ch    = make(chan *spotify.Client)
)

func main() {
	var client *spotify.Client
	http.HandleFunc("/callback", completeAuth)

	go func() {
		url := auth.AuthURL(state)
		fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)
		client = <-ch
		user, err := client.CurrentUser(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("You are logged in as:", user.ID)
	}()
	http.ListenAndServe(":8080", nil)

}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}
	client := spotify.New(auth.Client(r.Context(), tok))
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "Login Completed!")
	ch <- client
}

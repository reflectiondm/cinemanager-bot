package omdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type Rating struct {
	Source string `json:"Source"`
	Value  string `json:"Value"`
}

type Movie struct {
	Title      string   `json:"Title"`
	Year       string   `json:"Year"`
	Rated      string   `json:"Rated"`
	Released   string   `json:"Released"`
	Runtime    string   `json:"Runtime"`
	Genre      string   `json:"Genre"`
	Director   string   `json:"Director"`
	Writer     string   `json:"Writer"`
	Actors     string   `json:"Actors"`
	Plot       string   `json:"Plot"`
	Language   string   `json:"Language"`
	Country    string   `json:"Country"`
	Awards     string   `json:"Awards"`
	Poster     string   `json:"Poster"`
	Ratings    []Rating `json:"Ratings"`
	Metascore  string   `json:"Metascore"`
	ImdbRating string   `json:"imdbRating"`
	ImdbVotes  string   `json:"imdbVotes"`
	ImdbID     string   `json:"imdbID"`
	Type       string   `json:"Type"`
	DVD        string   `json:"DVD"`
	BoxOffice  string   `json:"BoxOffice"`
	Production string   `json:"Production"`
	Website    string   `json:"Website"`
	Response   string   `json:"Response"`
	Error      string   `json:"Error"`
}

func FetchMovieData(title string) (Movie, error) {
	apiKey := os.Getenv("OMDB_API_KEY")
	url := fmt.Sprintf("http://www.omdbapi.com/?apikey=%s&t=%s", apiKey, title)

	res, err := http.Get(url)
	if err != nil {
		return Movie{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return Movie{}, fmt.Errorf("API returned status code %d", res.StatusCode)
	}

	var result Movie
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return Movie{}, err
	}

	return result, nil
}

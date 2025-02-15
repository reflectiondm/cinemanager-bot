package storage

import (
	"cinemanager-bot/omdb"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

var (
	host     string
	port     = 5432
	user     string
	password string
	dbname   string

	db *sql.DB
)

func Init() {
	var err error

	host = os.Getenv("DB_HOST")
	user = os.Getenv("DB_USER")
	password = os.Getenv("DB_PASSWORD")
	dbname = os.Getenv("DB_NAME")
	port, _ = strconv.Atoi(os.Getenv("DB_PORT"))

	psqlconnection := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err = sql.Open("postgres", psqlconnection)
	CheckError(err)

	err = db.Ping()
	CheckError(err)

	runMigrations()

	fmt.Print("Successfully connected to database")
}

func DbDispose() {
	db.Close()
}

func SaveMovie(movie *omdb.Movie, messageId int) (int64, error) {
	insertStmt := `INSERT INTO movies (title, year, rated, released, runtime, genre, director, writer, actors, plot, language, country, awards, poster, metascore, imdb_rating, imdb_votes, imdb_id, type, dvd, box_office, production, website, response, error, message_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)`

	res, e := db.Exec(insertStmt, movie.Title, movie.Year, movie.Rated, movie.Released, movie.Runtime, movie.Genre, movie.Director, movie.Writer, movie.Actors, movie.Plot, movie.Language, movie.Country, movie.Awards, movie.Poster, movie.Metascore, movie.ImdbRating, movie.ImdbVotes, movie.ImdbID, movie.Type, movie.DVD, movie.BoxOffice, movie.Production, movie.Website, movie.Response, movie.Error, messageId)

	lastInsertId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastInsertId, e
}

func GetMovieByMessageId(messageId int) (*omdb.Movie, error) {
	var movie omdb.Movie

	row := db.QueryRow("SELECT * FROM movies WHERE message_id = $1", messageId)
	err := row.Scan(&movie.Title, &movie.Year, &movie.Rated, &movie.Released, &movie.Runtime, &movie.Genre, &movie.Director, &movie.Writer, &movie.Actors, &movie.Plot, &movie.Language, &movie.Country, &movie.Awards, &movie.Poster, &movie.Metascore, &movie.ImdbRating, &movie.ImdbVotes, &movie.ImdbID, &movie.Type, &movie.DVD, &movie.BoxOffice, &movie.Production, &movie.Website, &movie.Response, &movie.Error, &messageId)

	if err != nil {
		return nil, err
	}

	return &movie, nil
}

func DeleteMovieByMessageId(messageId int) error {
	_, err := db.Exec("DELETE FROM movies WHERE message_id = $1", messageId)
	return err
}

func GetMoviesBacklog() []omdb.Movie {
	var movies []omdb.Movie
	rows, err := db.Query("SELECT * FROM movies")
	CheckError(err)

	defer rows.Close()

	for rows.Next() {
		var movie omdb.Movie
		var messageId int
		err := rows.Scan(&movie.Title, &movie.Year, &movie.Rated, &movie.Released, &movie.Runtime, &movie.Genre, &movie.Director, &movie.Writer, &movie.Actors, &movie.Plot, &movie.Language, &movie.Country, &movie.Awards, &movie.Poster, &movie.Metascore, &movie.ImdbRating, &movie.ImdbVotes, &movie.ImdbID, &movie.Type, &movie.DVD, &movie.BoxOffice, &movie.Production, &movie.Website, &movie.Response, &movie.Error, &messageId)
		CheckError(err)

		movies = append(movies, movie)
	}

	return movies
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func runMigrations() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS movies (
		title TEXT,
		year TEXT,
		rated TEXT,
		released TEXT,
		runtime TEXT,
		genre TEXT,
		director TEXT,
		writer TEXT,
		actors TEXT,
		plot TEXT,
		language TEXT,
		country TEXT,
		awards TEXT,
		poster TEXT,
		metascore TEXT,
		imdb_rating TEXT,
		imdb_votes TEXT,
		imdb_id TEXT,
		type TEXT,
		dvd TEXT,
		box_office TEXT,
		production TEXT,
		website TEXT,
		response TEXT,
		error TEXT,
		message_id INTEGER
	)`)
	CheckError(err)

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_message_id ON movies (message_id)`)
	CheckError(err)
}

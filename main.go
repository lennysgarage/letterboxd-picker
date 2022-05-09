package main

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
)

type Movie struct {
	Title     string `json:"title"`
	Link      string `json:"movielink"`
	ImageLink string `json:"imagelink"`
}

// func userExists(username string) bool {
// 	file, err := os.Stat(fmt.Sprintf("watchlist-%s.csv", username))
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	modifiedTime := file.ModTime()

// 	if time.Now().After(modifiedTime.Add(time.Hour * 24))  {
// 		return true
// 	}

// 	return false
// }

func fetchWatchlist(username string) [][]string {

	var movies [][]string
	c := colly.NewCollector(
		colly.AllowedDomains("letterboxd.com"),
	)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting: ", r.URL.String())
	})

	c.OnHTML(".pagination", func(e *colly.HTMLElement) {
		nextPage := e.ChildAttr(".paginate-nextprev a.next", "href")
		c.Visit(e.Request.AbsoluteURL(nextPage))
	})

	// Find all movies in watchlist
	c.OnHTML(".poster-list li", func(e *colly.HTMLElement) {
		film := e.ChildAttr("div", "data-target-link")

		movie := Movie{}
		movie.Title = film[6 : len(film)-1]
		movie.Link = "https://letterboxd.com" + film

		row := []string{movie.Title, movie.Link}
		movies = append(movies, row)
	})

	c.Visit(fmt.Sprintf("https://letterboxd.com/%s/watchlist/page/1/", username))

	return movies
}

func chooseMovie(movies [][]string) Movie {
	randMovie := rand.Intn(len(movies))

	movie := Movie{}
	movie.Title = movies[randMovie][0]
	movie.Link = movies[randMovie][1]

	return movie
}

func fetchMoviePhoto(movieLink string) string {
	var movieImgLink string
	c := colly.NewCollector(
		colly.AllowedDomains("letterboxd.com"),
	)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting: ", r.URL.String())
	})

	c.OnHTML("#js-poster-col > section.poster-list.-p230.-single.no-hover.el.col > div", func(e *colly.HTMLElement) {
		temp := e.ChildAttr("img", "src")
		if temp != "" {
			movieImgLink = temp
		}
	})

	c.Visit(movieLink)
	return movieImgLink
}

func main() {
	router := gin.Default()
	router.GET("/api", func(c *gin.Context) {
		username := c.Query("username")
		movies := fetchWatchlist(username)
		randomMovie := chooseMovie(movies)
		randomMovie.ImageLink = fetchMoviePhoto(randomMovie.Link)

		c.JSON(http.StatusOK, randomMovie)
	})

	router.Run("localhost:8080")
}

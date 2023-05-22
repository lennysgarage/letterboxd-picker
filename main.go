package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

type Movie struct {
	Title     string `json:"title"`
	Link      string `json:"movielink"`
	ImageLink string `json:"imagelink"`
}

func fetchList(link string) []string {
	var movies []string
	c := colly.NewCollector(
		colly.AllowedDomains("letterboxd.com"),
		colly.Async(true),
	)

	c.OnError(func(e *colly.Response, err error) {
		log.Println("Something went wrong: ", err)
	})

	// Determines if a link to a list or a username.
	link = formatInput(link)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 4})

	extensions.RandomUserAgent(c)
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting:", r.URL.String())
	})

	// Fetch next page of watchlist
	c.OnHTML(".pagination", func(e *colly.HTMLElement) {
		nextPage := e.ChildAttr(".paginate-nextprev a.next", "href")
		c.Visit(e.Request.AbsoluteURL(nextPage))
	})

	// Find all movies in watchlist
	c.OnHTML(".poster-list li", func(e *colly.HTMLElement) {
		film := e.ChildAttr("div", "data-target-link")

		movie := Movie{}
		movie.Link = "https://letterboxd.com" + film

		movies = append(movies, movie.Link)
	})

	c.Visit(link)

	c.Wait()

	return movies
}

func formatInput(s string) string {
	if strings.HasPrefix(s, "http") {
		return s
	}

	return fmt.Sprintf("https://letterboxd.com/%s/watchlist/page/1/", s)
}

func chooseMovie(movies []string) Movie {
	randMovie := rand.Intn(len(movies))

	movie := Movie{}
	movie.Link = movies[randMovie]

	return movie
}

func fetchMovieInfo(movieLink string) (string, string) {
	var movieImgLink, movieTitle string
	c := colly.NewCollector(
		colly.AllowedDomains("letterboxd.com"),
		colly.Async(true),
	)

	c.OnError(func(e *colly.Response, err error) {
		log.Println("Something went wrong: ", err)
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting:", r.URL.String())
	})

	// Fetch movie poster title & link
	c.OnHTML("div.film-poster", func(e *colly.HTMLElement) {
		movieTitle = e.Attr("data-film-name")
		movieImgLink = e.ChildAttr("img", "src")
	})

	movieLink = "https://letterboxd.com/ajax/poster/" + strings.TrimPrefix(movieLink, "https://letterboxd.com/") + "std/230x345/"
	c.Visit(movieLink)

	c.Wait()

	return movieImgLink, movieTitle
}

func intersectLists(watchlist []string, numUsers int) []string {
	intersection := make([]string, 0)
	hash := make(map[string]int)

	for _, movie := range watchlist {
		hash[movie]++
	}

	for movie, count := range hash {
		if count == numUsers {
			intersection = append(intersection, movie)
		}
	}
	return intersection
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	var wg sync.WaitGroup

	router := gin.Default()
	router.Use(cors.Default())
	router.GET("/api", func(c *gin.Context) {
		usernames := c.QueryArray("src")
		intersection := c.Query("i")

		var movieList []string
		if len(usernames) > 1 && intersection == "true" {
			var allMovies []string
			// Fetch all user's watchlists
			for _, username := range usernames {
				wg.Add(1)

				go func(username string) {
					defer wg.Done()
					movies := fetchList(username)
					allMovies = append(allMovies, movies...)
				}(username)
			}
			wg.Wait()
			// Create intersected watchlist
			movieList = intersectLists(allMovies, len(usernames))
		} else { // Union
			// Pick a random user
			user := usernames[rand.Intn(len(usernames))]
			// Fetch single user's watchlist, equivalent to union since randomness
			movieList = fetchList(user)
		}

		// Return movie
		if len(movieList) != 0 {
			randomMovie := chooseMovie(movieList)
			randomMovie.ImageLink, randomMovie.Title = fetchMovieInfo(randomMovie.Link)
			c.JSON(http.StatusOK, randomMovie)
		} else {
			c.Status(http.StatusNotFound)
		}
	})

	err := router.Run(":" + port)
	if err != nil {
		log.Fatal("Server crashed unexpectedly ", err)
	}
}

package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
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

func fetchWatchlist(username string) []string {
	var movies []string
	c := colly.NewCollector(
		colly.AllowedDomains("letterboxd.com"),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 4})
	extensions.RandomUserAgent(c)
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting: ", r.URL.String())
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

	c.Visit(fmt.Sprintf("https://letterboxd.com/%s/watchlist/page/1/", username))

	c.Wait()
	return movies
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
	)
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting: ", r.URL.String())
	})

	// Fetch movie title
	c.OnHTML("meta[property='og:title']", func(e *colly.HTMLElement) {
		movieTitle = e.Attr("content")
	})

	// Fetch movie poster link
	c.OnHTML("#js-poster-col > section.poster-list.-p230.-single.no-hover.el.col > div", func(e *colly.HTMLElement) {
		temp := e.ChildAttr("img", "src")
		if temp != "" {
			movieImgLink = temp
		}
	})

	c.Visit(movieLink)
	return movieImgLink, movieTitle
}

func intersectWatchlists(watchlistA, watchlistB []string) []string {
	intersection := make([]string, 0)
	hash := make(map[string]bool)

	for _, movie := range watchlistA {
		hash[movie] = true
	}

	for _, movie := range watchlistB {
		if hash[movie] {
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
		usernames := c.QueryArray("username")
		intersection := c.Query("i")

		// atm limited to 2 users
		var movieList []string
		if len(usernames) == 2 && intersection == "true" {
			var allMovies [][]string
			// Fetch both user's watchlists
			for _, username := range usernames {
				wg.Add(1)

				go func(username string) {
					defer wg.Done()
					movies := fetchWatchlist(username)
					allMovies = append(allMovies, movies)
				}(username)
			}
			wg.Wait()
			// Create intersected watchlist
			movieList = intersectWatchlists(allMovies[0], allMovies[1])
		} else { // Union
			// Pick a random user
			user := usernames[rand.Intn(len(usernames))]
			// Fetch single user's watchlist, equivalent to union since randomness
			movieList = fetchWatchlist(user)
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

	router.Run(":" + port)
}

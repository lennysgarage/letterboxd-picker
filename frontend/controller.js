document.getElementsByTagName('form').item(0).addEventListener('submit', sendRequest());

// document.getElementsByName('unionAndIntersection').forEach((btn) => btn.addEventListener('onClick', func(btn)))

hideLoading();

function sendRequest() {
    return async function(e) {
        e.preventDefault();
        let usernames = document.getElementById('username').value.trim().split(" ", 6); // limit to 6 users
        if (usernames?.length > 0 && usernames[0] !== '') {
            showLoading();
            // development: 
            // let urlString = "http://localhost:8080/api?";
            // production: 
            let urlString = "https://letterboxd-picker-api.herokuapp.com/api?";    
            
            usernames.forEach((username) => {
                urlString += "username=" + username + "&";
            })
            if (getType() === "intersection") {
                urlString += "i=true";
            }

            console.log("url: " + urlString)
            const response = await fetch(urlString)
            hideLoading();

            let movie = document.getElementById("movie-container");
            if (response.status === 200) {
                const content = await response.json();
                movie.innerHTML = `<a 
                    class="movie-poster" 
                    href="${content.movielink}" 
                    style="background-image: url('${content.imagelink}');"></a>
                    <p id="movieTitle">You should watch <a id="movieTitleLink" href="${content.movielink}">${content.title}</a></p>`
            } else {
                movie.innerHTML = `<p id="missingMovie">Sorry that watchlist does not exist.</p>
                <img class="not-found" src="404-not-found.gif" alt="20th century fox intro, replaced with 404 Not found error.">`;
            }
        }
    }
}

function showLoading() {
    document.getElementById('submitButton').innerHTML = '<span id="spinner" class="spinner-border text-light spinner-border-sm" role="status" aria-hidden="true"></span>Loading...';
}

function hideLoading() {
    document.getElementById('submitButton').innerHTML = 'SUBMIT';
}

function getType() {
    if (document.querySelector('input[id="intersection"]:checked') !== null) {
        return "intersection";
    }
    return "union";
}
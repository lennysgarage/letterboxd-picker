document.getElementsByTagName('form').item(0).addEventListener('submit', sendRequest());

hideLoading();

function sendRequest() {
    return async function(e) {
        e.preventDefault();
        let username = document.getElementById('username').value;
        if (username !== '') {
            showLoading();
            // development: http://localhost:8080/api?username=${username}
            // production: https://letterboxd-picker-api.herokuapp.com/api?username=${username}
            const response = await fetch(`https://letterboxd-picker-api.herokuapp.com/api?username=${username}`);
            hideLoading();
            let movie = document.getElementById("movie-container");
            if (response.status === 200) {
                const content = await response.json();
                let movieTitle = content.title.replace(/-/g, " ");
                movie.innerHTML = `<a 
                    class="movie-poster" 
                    href="${content.movielink}" 
                    style="background-image: url('${content.imagelink}');"></a>
                    <p id="movieTitle">You should watch <a id="movieTitleLink" href="${content.movielink}">${movieTitle.charAt(0).toUpperCase() + movieTitle.slice(1)}</a></p>`
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
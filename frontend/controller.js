document.getElementsByTagName('form').item(0).addEventListener('submit', sendRequest());

hideLoading();

function sendRequest() {
    return async function(e) {
        e.preventDefault();
        let username = document.getElementById('username').value;
        if (username !== '') {
            showLoading();
            const response = await fetch(`http://localhost:8080/api?username=${username}`);
            hideLoading();
            if (response.status === 200) {
                const content = await response.json();
                let movie = document.getElementById("movie-container");
                movie.innerHTML = `<a 
                    class="movie-poster" 
                    href="${content.movielink}" 
                    style="background-image: url('${content.imagelink}');"></a>`
            } else {
                console.log(response.status)
            }
        }
    }
}

function showLoading() {
    document.getElementById('spinner').style.display = 'block';
}

function hideLoading() {
    document.getElementById('spinner').style.display = 'none';
}
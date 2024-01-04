function search() {
  const query = searchBar.value
  const start = Date.now()
  fetch(`${location.origin}/api/searchResults?prompt=${encodeURIComponent(query)}`)
    .catch(onError)
    .then(res => res.json())
    .then(data => {
      if (data == null || data.length == 1) {
        onError()
      } else {
        setSearchState(false)
        location.href = `/searchResults.html#${btoa(encodeURIComponent(JSON.stringify({timings: Date.now() - start, results: data, query: query})))}`
      }
    })
  setSearchState(true)
}

document.addEventListener("DOMContentLoaded", () => {
  window.searchBar = document.getElementById("searchbar")
  window.searchButton = document.getElementById("search")

  searchButton.addEventListener("click", search)

  searchBar.addEventListener("input", () => {
    if (searchBar.value.length > 0) {
    searchButton.classList.remove("not-allowed");
  } else {
    searchButton.classList.add("not-allowed")
  }
  })
})

document.addEventListener("keydown", e => e.code === "Enter" && !searchBar?.disabled ? search() : null)

function setSearchState(to) {
  searchBar.disabled = to
  searchButton.disabled = to
}

function onError() {
  alert("An error has occured! Your request may not be appropriate for the AI. Try something else")
  setSearchState(false)
}


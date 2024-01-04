const data = JSON.parse(decodeURIComponent(atob(location.hash.slice(1))))

document.addEventListener("keydown", e => e.code === "Enter" && !searchBar?.disabled ? search() : null)

function setSearchState(to) {
  searchBar.disabled = to
  searchButton.disabled = to
}

function onError() {
  alert("An error has occured! Your request may not be appropriate for the AI. Try something else")
  setSearchState(false)
}

function Element(tag, data={}) {
  return Object.assign(document.createElement(tag), data)
}

function goToWebpage(Name, Description) {
  location.href = `/api/webPage/?name=${encodeURIComponent(Name)}&description=${encodeURIComponent(Description)}`
}

function makeSearchResult({Name, Description}) {
  const container = Element("div", {className: "searchResult"})
  container.append(
    Element("h2", {textContent: Name, onclick: () => goToWebpage(Name, Description)}),
    Element("p", {textContent: Description})
  )
  return container
}


document.addEventListener("DOMContentLoaded", () => {
  window.found = document.getElementById("found")
  window.timings = document.getElementById("timings")

  found.append(...data.results.map(makeSearchResult))
  timings.textContent = `Found ${data.results.length} results in ${data.timings / 1000}s`
  document.getElementById("search").textContent = data.query
})
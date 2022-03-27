document.addEventListener("DOMContentLoaded", (event) => {
  setFilters();
  // filter set event listeners to remove section and book options as filters are set

  document.getElementById("testamentFilter").onchange = updateSections;
  document.getElementById("sectionFilter").onchange = updateBooks;

  // set pagination buttons
  // make sure we are on the search page and not chapter browsing
  if (!!document.getElementById("results")) {
    createPagination();
  }
});

function setFilters() {
  // persist search input value
  let params = new URL(document.location).searchParams;
  let search = params.get("search");
  let testamentFilter = params.get("testamentFilter");
  let sectionFilter = params.get("sectionFilter");
  let bookFilter = params.get("bookFilter");
  let fakeFilter = params.get("fakeFilter");

  let queries = { search, testamentFilter, sectionFilter, bookFilter };

  for (const key in queries) {
    let q = queries[key];
    let sq = localStorage.getItem(key);
    if (q == "none") {
      localStorage.removeItem(key);
    } else if (q && (!sq || sq != q)) {
      localStorage.setItem(key, q);
    }
    if (!q && sq) {
      queries[key] = sq;
    }
  }

  if (!!document.getElementById("results")) {
    window.history.pushState({}, document.title, "/");
  }
  // set the filters on page load
  if (queries.testamentFilter) {
    document.getElementById("testamentFilter").value = queries.testamentFilter;
  }

  if (queries.sectionFilter) {
    document.getElementById("sectionFilter").value = queries.sectionFilter;
  }

  if (queries.bookFilter) {
    // wait for books api
    setTimeout(function () {
      document.getElementById("bookFilter").value = queries.bookFilter;
    }, 500);
  }

  if (queries.search) {
    document.getElementById("search").value = queries.search;
  }
}

function createPagination() {
  let moreRes = document.getElementById("next-results");
  let prevRes = document.getElementById("prev-results");
  let resultCount = Number(document.getElementById("results").dataset.count);

  let currentPage = Number(params.get("p")) ? Number(params.get("p")) : 1;
  let prevPage;
  let nextPage;

  if (currentPage <= 1) {
    nextPage = 2;
    prevRes.remove();
    if (resultCount / (currentPage * 50) < 1) {
      moreRes.remove();
    }
  } else if (resultCount / (currentPage * 50) < 1) {
    prevPage = currentPage - 1;
    moreRes.remove();
  } else {
    nextPage = currentPage + 1;
    prevPage = currentPage - 1;
  }

  params.set("p", nextPage);
  moreRes.href = document.location.origin + "?" + params;

  params.set("p", prevPage);
  prevRes.href = document.location.origin + "?" + params;
}

function updateSections(e) {
  // look at every section, check data-testament for matching value e.target.value
  let v = e.target.value;

  // cant just remove the nodes because i have to be able to change the filter back
  // loop through sections
  // let opts = document.getElementById("sectionFilter").children;
  // console.log(opts);
  // for (i = 0; i < opts.length; i++) {
  //   if (opts[i].dataset.testament != v) opts[i].remove();
  // }
}
function updateBooks(e) {
  // look at every book, check data-section for matching value e.target.value
  console.log(e);
}

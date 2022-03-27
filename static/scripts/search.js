document.addEventListener("DOMContentLoaded", (event) => {
  let params = new URL(document.location).searchParams;

  let queries = setFilters(params);
  // filter set event listeners to remove section and book options as filters are set
  document.getElementById("testamentFilter").onchange = handleSelectChange;
  document.getElementById("sectionFilter").onchange = handleSelectChange;
  document.getElementById("searchForm").onsubmit = handleSearch;

  // set pagination buttons
  if (!!document.getElementById("results")) {
    createPagination(params);
  }
});

function setFilters(params) {
  // persist search input value
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
  // This should no longer be needed. Should remove after testing
  // if (!!document.getElementById("results")) {
  //   window.history.pushState({}, document.title, "/");
  // }
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
    }, 200);
  }

  if (queries.search) {
    document.getElementById("search").value = queries.search;
  }

  updateSections(queries.testamentFilter);
  setTimeout(() => {
    updateBooks(queries.sectionFilter);
  }, 600);
  return queries;
}

function createPagination(params) {
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

function updateSections(value) {
  let v = value;

  let opts = document.getElementById("sectionFilter").children;
  for (i = 0; i < opts.length; i++) {
    if (v == "none" || opts[i].dataset.testament == v) {
      opts[i].removeAttribute("disabled");
    } else {
      opts[i].setAttribute("disabled", true);
    }
  }
  opts[0].removeAttribute("disabled");
}
function updateBooks(value) {
  let v = value;

  let opts = document.getElementById("bookFilter").children;
  for (i = 0; i < opts.length; i++) {
    if (v == "none" || opts[i].dataset.section != v) {
      opts[i].setAttribute("disabled", true);
    } else {
      opts[i].removeAttribute("disabled");
    }
  }
  opts[0].removeAttribute("disabled");
}

function handleSelectChange(e) {
  if (e.target.id == "testamentFilter") {
    updateSections(e.target.value);
    document.getElementById("sectionFilter").value = "none";
    document.getElementById("bookFilter").value = "none";
  } else if (e.target.id == "sectionFilter") {
    updateBooks(e.target.value);
    document.getElementById("bookFilter").value = "none";
  }
}

function handleSearch(e) {
  e.preventDefault();

  let search = document.getElementById("search").value;
  let testamentFilter = document.getElementById("testamentFilter").value;
  let sectionFilter = document.getElementById("sectionFilter").value;
  let bookFilter = document.getElementById("bookFilter").value;

  window.location.href = `${window.location.origin}?search=${search}${
    testamentFilter != "none" ? `&testamentFilter=${testamentFilter}` : ""
  }${sectionFilter != "none" ? `&sectionFilter=${sectionFilter}` : ""}${
    bookFilter != "none" ? `&bookFilter=${bookFilter}` : ""
  }`;
}

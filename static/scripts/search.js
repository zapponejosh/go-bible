document.addEventListener("DOMContentLoaded", (event) => {
  // persist search input value
  let params = new URL(document.location).searchParams;
  let search = params.get("search");
  let testamentFilter = params.get("testamentFilter");
  let sectionFilter = params.get("sectionFilter");
  let bookFilter = params.get("bookFilter");

  let paramObj = { search, testamentFilter, sectionFilter, bookFilter };

  // remove unused filters from params
  for (const key in paramObj) {
    if (!paramObj[key]) {
      params.delete(key);
    }
  }
  let paramsString = params.toString();
  //update url
  window.history.pushState({}, document.title, "/?" + paramsString);

  // set the filters
  if (testamentFilter) {
    let opt = document.querySelector(
      `#${"testamentFilter"} > option[value]:is([value='${testamentFilter}'])`
    );
    opt.setAttribute("selected", true);
  }

  if (sectionFilter) {
    let opt = document.querySelector(
      `#${"sectionFilter"} > option[value]:is([value='${sectionFilter}'])`
    );
    opt.setAttribute("selected", true);
  }

  if (bookFilter) {
    setTimeout(function () {
      let opt = document.querySelector(
        `#${"bookFilter"} > option[value]:is([value='${bookFilter}'])`
      );
      opt.setAttribute("selected", true);
    }, 500);
  }

  // set ref urls and input value
  if (search) {
    let references = document.querySelectorAll(".reference");
    references.forEach((refNode) => {
      refNode.href = refNode.href + "&search=" + search;
    });
    let searchInput = document.getElementById("search");
    searchInput.value = search;
  }

  if (!!document.getElementById("results")) {
    // set pagination buttons
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
});

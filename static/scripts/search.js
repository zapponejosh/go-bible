document.addEventListener("DOMContentLoaded", (event) => {
  let params = new URL(document.location).searchParams;
  let search = params.get("search");
  if (search) {
    let references = document.querySelectorAll(".reference");
    references.forEach((refNode) => {
      refNode.href = refNode.href + "&search=" + search;
    });
    let searchInput = document.getElementById("search");
    searchInput.value = search;
  }
});

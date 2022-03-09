document.addEventListener("DOMContentLoaded", (event) => {
  let params = new URL(document.location).searchParams;
  let ref = params.get("ref");
  if (ref) {
    let referringVerse = document.querySelector(`[data-verse='${ref}']`);
    referringVerse.className = "highlight";
    referringVerse.scrollIntoView({ block: "center" });
  }
});

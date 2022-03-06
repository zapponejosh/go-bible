// chapter select form
let bookSelect = document.getElementById("bookSelect");
let chapterSelect = document.getElementById("chapterSelect");
let form = document.getElementById("chapterForm");
// get options
fetch("/api/books")
  .then((res) => res.json())
  .then((data) => {
    data.forEach((book) => {
      let bookOpt = document.createElement("option");
      bookOpt.text = book.long_name;
      bookOpt.id = book.short_name;
      bookOpt.value = book.short_name;
      bookOpt.dataset.chapters = book.chapters;
      bookOpt.dataset.section = book.section;

      bookSelect.appendChild(bookOpt);
    });

    let chosenBook = bookSelect.firstElementChild;
    console.log(chosenBook.dataset.chapters);
    for (i = chosenBook.dataset.chapters; i > 0; i--) {
      let chOpt = document.createElement("option");
      chOpt.text = i;
      chOpt.value = i;

      chapterSelect.prepend(chOpt);
    }
  })
  .catch((err) => alert(err));

bookSelect.onchange = handleBookChange;

// form submission
form.onsubmit = submit;

function submit(event) {
  event.preventDefault();
  let book = bookSelect.value;
  let ch = chapterSelect.value;
  window.location = `/${book}/${ch}`;
}

function handleBookChange(event) {
  let book = bookSelect.value;
  let el = document.getElementById(book);
  // remove existing options
  while (chapterSelect.firstChild) {
    chapterSelect.removeChild(chapterSelect.firstChild);
  }

  for (i = el.dataset.chapters; i > 0; i--) {
    let chOpt = document.createElement("option");
    chOpt.text = i;
    chOpt.value = i;

    chapterSelect.prepend(chOpt);
  }
}

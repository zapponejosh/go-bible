document.addEventListener("DOMContentLoaded", (event) => {
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
        bookOpt.className = book.section;

        bookSelect.appendChild(bookOpt);
      });

      // current state
      let currentLoc = window.location.pathname;
      currentBook = currentLoc.split("/")[1];
      currentChapter = Number(currentLoc.split("/")[2]);
      let bookData = data.find((book) => book.short_name == currentBook);

      // if not on chapter page
      if (!bookData || !currentChapter) {
        let chosenBook = bookSelect.firstElementChild;
        for (i = chosenBook.dataset.chapters; i > 0; i--) {
          let chOpt = document.createElement("option");
          chOpt.text = i;
          chOpt.value = i;

          chapterSelect.prepend(chOpt);
        }
        // else on chapter page
      } else {
        // set dropdown to current book
        let bookOpt = document.getElementById(currentBook);
        bookOpt.setAttribute("selected", true);
        for (i = bookOpt.dataset.chapters; i > 0; i--) {
          let chOpt = document.createElement("option");
          chOpt.text = i;
          chOpt.value = i;

          chapterSelect.prepend(chOpt);
        }

        // set selected chapter
        let chapterOpt = document.querySelector(
          `#chapterSelect > option[value]:is([value='${currentChapter}'])`
        );
        chapterOpt.setAttribute("selected", true);
        // set buttons for previous and next chatpers
        let nextBtn = document.getElementById("next");
        let previousBtn = document.getElementById("previous");

        // previous button
        if (currentChapter > 1) {
          previousBtn.href = `/${currentBook}/${currentChapter - 1}`;
        } else {
          previousBtn.href = "";
          previousBtn.className = "isDisabled";
        }
        // next button
        let chapterCount = Number(bookData.chapters);
        if (currentChapter < chapterCount) {
          nextBtn.href = `/${currentBook}/${currentChapter + 1}`;
        } else {
          nextBtn.href = "";
          nextBtn.className = "isDisabled";
        }
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
      if (i == 1) {
        chOpt.selected = true;
      }

      chapterSelect.prepend(chOpt);
    }
  }
});

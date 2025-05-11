console.log(1);

function filterCategories() {
  var checkboxes = document.querySelectorAll(
    '.dropdown-content input[type="checkbox"]',
  );
  const selectedCats = [];
  checkboxes.forEach((checkbox) => {
    if (checkbox.checked) {
      selectedCats.push(checkbox.value);
    }
  });

  const transactions = document.querySelectorAll("[data-category]");
  for (const element of transactions) {
    if (!selectedCats.includes(element.dataset.category)) {
      element.style.display = "none";
    } else {
      element.style.display = "table-row";
    }
  }
}

function showDropdown() {
  const dropdown = document.querySelector(".dropdown-content");
  dropdown.style.display = "block";
}

document.addEventListener("click", (event) => {
  const dropdown = document.querySelector(".dropdown");
  const dropdownContent = dropdown.querySelector(".dropdown-content");
  if (!dropdown.contains(event.target)) {
    dropdownContent.style.display = "none";
    filterCategories();
  }
});

function updateSort(sortCol) {
  if ('URLSearchParams' in window) {
    const searchParams = new URLSearchParams(window.location.search);
    if (searchParams.get("sortCol") == sortCol) {
      // already sorting on this col, so reverse direction
      let oldDir = searchParams.get("sortDir");
      let newDir = oldDir != "desc" ? "desc" : "asc";
      searchParams.set("sortDir", newDir);
    } else {
      searchParams.set("sortCol", sortCol);
      searchParams.set("sortDir", "asc");
    }
    window.location.search = searchParams.toString();
  }
}

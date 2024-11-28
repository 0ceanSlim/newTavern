 // Immediate execution of theme logic
 (function () {
  function setTheme(theme) {
    document.documentElement.setAttribute("data-theme", theme);
    localStorage.setItem("theme", theme);
  }

  // Initialize dropdown toggle
  window.toggleDropdown = function (id) {
    const dropdown = document.getElementById(id);
    dropdown.classList.toggle("hidden");
  };

  // Set theme based on saved preference
  const savedTheme = localStorage.getItem("theme");
  if (savedTheme) {
    setTheme(savedTheme);
  }

  // Add click event listeners for swatches
  document.querySelectorAll(".swatch").forEach((button) => {
    button.addEventListener("click", () => {
      const newTheme = button.dataset.theme;
      setTheme(newTheme);

      // Close the dropdown
      const dropdown = document.getElementById("themeDropdown");
      dropdown.classList.add("hidden");
    });
  });
})();
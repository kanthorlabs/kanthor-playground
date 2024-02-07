/* globals bootstrap:false, Prism:false */

(function () {
  "use strict";

  // Add/remove `.navbar-transparent` on scroll; should probably be throttled later
  function addNavbarTransparentClass() {
    const navBarElement = document.querySelector("#home > .navbar");

    if (!navBarElement) {
      return;
    }

    window.addEventListener("scroll", () => {
      const scroll = document.documentElement.scrollTop;

      if (scroll > 50) {
        navBarElement.classList.remove("navbar-transparent");
      } else {
        navBarElement.classList.add("navbar-transparent");
      }
    });
  }

  // Toggle light and dark themes
  function toggleThemeMenu() {
    let themeMenu = document.querySelector("#theme-menu");

    if (!themeMenu) return;

    document.querySelectorAll("[data-bs-theme-value]").forEach((value) => {
      value.addEventListener("click", () => {
        const theme = value.getAttribute("data-bs-theme-value");
        document.documentElement.setAttribute("data-bs-theme", theme);
      });
    });
  }

  addNavbarTransparentClass();

  toggleThemeMenu();

  // Prevent empty `a` elements or `submit` buttons from navigating away
  const targets = document.querySelectorAll('[href="#"], [type="submit"]');

  for (const element of targets) {
    element.addEventListener("click", (event) => {
      event.preventDefault();
    });
  }

  // Initialize popovers
  const popoverElements = document.querySelectorAll(
    '[data-bs-toggle="popover"]'
  );

  for (const popover of popoverElements) {
    new bootstrap.Popover(popover); // eslint-disable-line no-new
  }

  // Initialize tooltips
  const tooltipElements = document.querySelectorAll(
    '[data-bs-toggle="tooltip"]'
  );

  for (const tooltip of tooltipElements) {
    new bootstrap.Tooltip(tooltip); // eslint-disable-line no-new
  }
})();

function age(date) {
  var seconds = Math.floor((new Date() - date) / 1000);
  var interval = seconds / 31536000;
  if (interval > 1) {
    return Math.floor(interval) + " years ago";
  }
  interval = seconds / 2592000;
  if (interval > 1) {
    return Math.floor(interval) + " months ago";
  }
  interval = seconds / 86400;
  if (interval > 1) {
    return Math.floor(interval) + " days ago";
  }
  interval = seconds / 3600;
  if (interval > 1) {
    return Math.floor(interval) + " hours ago";
  }
  interval = seconds / 60;
  if (interval > 1) {
    return Math.floor(interval) + " minutes ago";
  }
  return Math.floor(seconds) + " seconds ago";
}

function replaceAt(text, anchor, replacement) {
  index = text.indexOf(anchor) + anchor.length;
  return (
    text.substring(0, index) +
    replacement +
    text.substring(index + replacement.length)
  );
}

function pretty(data) {
  try {
    return JSON.stringify(JSON.parse(data), null, 2);
  } catch {
    return data;
  }
}

function prepare() {
  const elements = document.querySelectorAll(".kanthor-message");
  for (let element of elements) {
    const id = element.getAttribute("data-message-id");
    const timestamp = element.getAttribute("data-message-timestamp");
    const headers = element.getAttribute("data-message-headers");
    const body = element.getAttribute("data-message-body");

    const pass = age(new Date(timestamp));
    element.querySelector(".kanthor-message-timestamp").textContent = pass;

    element.addEventListener("click", (event) => {
      event.preventDefault();

      const details = document.getElementById("kanthor-message-details");
      details.querySelector(".card-title").textContent = id;
      details.querySelector(".card-subtitle").textContent = pass;

      details.querySelector(".kanthor-message-headers").innerHTML = `
        <h5>Headers</h5>
        <pre>
          <code class="language-json">${pretty(headers)}</code>
        </pre>`;
      details.querySelector(".kanthor-message-body").innerHTML = `
        <h5>Body</h5>
        <pre>
          <code class="language-json">${pretty(body)}</code>
        </pre>`;

      // pre-highlight
      hljs.highlightAll();
    });
  }
  hljs.addPlugin(
    new CopyButtonPlugin({
      hook: (text, el) => {
        if (el.classList.contains("language-json")) return text;

        const updated = replaceAt(
          text.replace(/  /g, "").slice("\n".length),
          "Idempotency-Key: ",
          uuidv4()
        );
        return updated;
      },
    })
  );
  hljs.highlightAll();
}

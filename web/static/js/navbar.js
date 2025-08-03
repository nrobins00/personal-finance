const navbarNodes = Array.from(document.getElementsByClassName("topnav")[0].children)

navbarNodes.forEach(element => {
    if (element.href === window.location.href) {
        element.classList.add("active")
    }
});
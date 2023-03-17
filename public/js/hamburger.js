// hamburger menu


const menu = document.querySelector(".toggle");
const menuTogle = document.querySelector(".toggle input");
const nav = document.querySelector("nav ul");
let isOpen = false;

menu.addEventListener("click", () => {
    if (isOpen) {
        menu.classList.remove("active");
        isOpen = false;
    } else {
        menu.classList.add("active");
        isOpen = true;
    }
})


menuTogle.addEventListener("click", function () {
    nav.classList.toggle('slide');
})


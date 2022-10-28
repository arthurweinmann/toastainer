var isCollapsed = localStorage.getItem("sidemenu-collapsed");

function initSidemenu() {
    var sidemenu = document.querySelector(".sidemenu");
    var iconExpandCollapse = document.querySelector(".icon-expand-collapse");
    var btnSidemenu = document.querySelectorAll(".sidemenu__header button");
    var labelItemSidemenu = document.querySelectorAll(".sidemenu__item span");
    var iconItemSidemenu = document.querySelectorAll(".sidemenu__cont-icon");
    var logoText = document.querySelector(".header__logo-img-text");
    var contentTitle = document.querySelector(".content__wrapper-title");
    var wrapperSlider = document.querySelectorAll(".wrapper-slider");
    var btnMenuMobile = document.getElementById("btn-menu-mobile");
    var menuMobile = document.querySelector(".sidemenu__content-mobile");
    let timeoutMenuMobile;
    let timeoutCollapse = 60;

    if (isCollapsed === "true") {
        collapseSidemenu();
    }

    iconExpandCollapse.addEventListener("click", function () {
        if (sidemenu.classList.contains("collapsed")) {
            localStorage.setItem("sidemenu-collapsed", null);
            expandSidemenu();
        }
        else {
            localStorage.setItem("sidemenu-collapsed", "true");
            collapseSidemenu();
        }
    });

    function collapseSidemenu() {
        sidemenu.classList.add("collapsed");
        logoText.classList.add("hidden");

        btnSidemenu.forEach(btn => {
            btn.classList.add("hidden");
        });

        labelItemSidemenu.forEach(labelItem => {
            labelItem.classList.add("hidden");
        });

        iconItemSidemenu.forEach((iconItem, i) => {
            setTimeout(() => {
                iconItem.classList.add("collapsed");
            }, timeoutCollapse * (i + 1));
        });

        if (contentTitle) {
            contentTitle.classList.add("collapsed");
        }

        if (wrapperSlider) {
            wrapperSlider.forEach(slider => {
                slider.classList.add("collapsed");
            });
        }
    }

    function expandSidemenu() {
        if (wrapperSlider) {
            wrapperSlider.forEach(slider => {
                slider.classList.remove("collapsed");
            });
        }
        sidemenu.classList.remove("collapsed");
        logoText.classList.remove("hidden");

        btnSidemenu.forEach(btn => {
            btn.classList.remove("hidden");
        });

        labelItemSidemenu.forEach((labelItem, i) => {
            setTimeout(() => {
                labelItem.classList.remove("hidden");
            }, timeoutCollapse * (i + 1));
        });

        iconItemSidemenu.forEach((iconItem, i) => {
            setTimeout(() => {
                iconItem.classList.remove("collapsed");
            }, timeoutCollapse * (i + 1));
        });

        if (contentTitle) {
            contentTitle.classList.remove("collapsed");
        }
    }

    btnMenuMobile.addEventListener("click", function () {
        clearTimeout(timeoutMenuMobile);
        
        if (menuMobile.classList.contains("show")) {
            menuMobile.classList.remove("show");

            timeoutMenuMobile = setTimeout(() => {
                menuMobile.style.display = "none";
            }, 300);
        }
        else {
            menuMobile.style.display = "block";

            timeoutMenuMobile = setTimeout(() => {
                menuMobile.classList.add("show");
            }, 50);
        }
    });
}


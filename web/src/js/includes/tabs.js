var tabsLink = document.querySelectorAll(".tab__link");
var tabsContent = document.querySelectorAll(".tab__content");

tabsLink.forEach(link => {
    link.addEventListener("click", function () {
        let value = this.getAttribute("value");
        let tabContentToShow = document.getElementById("tabcontent-" + value);

        for (let i = 0; i < tabsContent.length; i++) {
            tabsContent[i].classList.remove("show");
        }

        for (let i = 0; i < tabsLink.length; i++) {
            tabsLink[i].classList.remove("active");
        }

        tabContentToShow.classList.add("show");
        this.classList.add("active");
    });
});

function showTab(idTab) {
    let tabContentToShow = document.getElementById("tabcontent-" + idTab);
    let tabLinkToActive = document.getElementById("tablink-" + idTab);

    for (let i = 0; i < tabsContent.length; i++) {
        tabsContent[i].classList.remove("show");
    }

    for (let i = 0; i < tabsLink.length; i++) {
        tabsLink[i].classList.remove("active");
    }

    tabContentToShow.classList.add("show");
    tabLinkToActive.classList.add("active")
}
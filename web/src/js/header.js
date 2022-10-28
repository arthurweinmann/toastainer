import "local://includes/routes/toaster.js";
import "local://includes/routes/user.js";
import "local://includes/toastTile.js";

var timeoutSearchToaster;
var gListToastersForSearch = null;

var headerAccount = document.querySelector(".header-account__profile");
var inpSearchToaster = document.getElementById("inp-search-toaster");
var searchScreen = document.querySelector(".search-screen");
var contSearch = document.querySelector(".cont-search");
var btnCloseSearchScreen = document.querySelector(".close-search-screen");
var tileResultElem = document.querySelector(".section-search-toasters .tile__count");
var emptyResultToasterElem = document.querySelector(".toastTile__empty-from-search");
var gSearchToasterCount = 0;

var blockSearchToasters = document.querySelector(".section-search-toasters");

if (headerAccount) {
    headerAccount.addEventListener("click", function () {
        window.location = "/account";
    });
}

if (btnCloseSearchScreen) {
    btnCloseSearchScreen.addEventListener("click", function () {
        closeSearchScreen();
    });
}

if (inpSearchToaster) {
    inpSearchToaster.addEventListener("keyup", function (e) {
        var toastTilesSearch;
        let searchToaster = e.target.value;

        clearTimeout(timeoutSearchToaster);

        if (searchToaster.length === 0) {
            document.documentElement.style.overflow = "auto";
            searchScreen.classList.remove("show");
            timeoutSearchToaster = setTimeout(() => {
                searchScreen.style.display = "none";
            }, 300);

        }
        else {
            document.documentElement.style.overflow = "hidden";
            searchScreen.style.display = "block";
            timeoutSearchToaster = setTimeout(() => {
                searchScreen.classList.add("show");
                contSearch.classList.add("show");
            }, 50);
        }

        if (!gListToastersForSearch) {
            handleDisplaySearchToasters();
        }
        else {
            toastTilesSearch = document.querySelectorAll(".toastTile__item.from-search");

            if (e.target.value.length > 0) {
                toastTilesSearch.forEach(toasterSearch => toasterSearch.classList.remove("hidden"));
                toastTilesSearch.forEach(toasterSearch => {
                    let toasterName = toasterSearch.getAttribute("data-toastername");
                    if (!toasterName.toLowerCase().includes(e.target.value.trim().toLowerCase())) {
                        toasterSearch.classList.add("hidden");
                    }
                });

                var toastersNoHidden =  document.querySelectorAll(".toastTile__item.from-search:not(.hidden)");
                gSearchToasterCount = toastersNoHidden ? toastersNoHidden.length : 0;
                tileResultElem.textContent = gSearchToasterCount;

                if (gSearchToasterCount === 0) {
                    emptyResultToasterElem.classList.add("show");
                }
                else if (emptyResultToasterElem.classList.contains("show")) {
                    emptyResultToasterElem.classList.remove("show");
                }
            }
        }

    });
}

function handleDisplaySearchToasters() {
    handleRenderToasters(blockSearchToasters, {
        isFromSearch: true,
        deleteCallback: handlePropagateMyToasters,
        endCallback: () => {
            toastTilesSearch = document.querySelectorAll(".toastTile__item.from-search");

            if (inpSearchToaster.value.length > 0) {
                toastTilesSearch.forEach(toasterSearch => toasterSearch.classList.remove("hidden"));
                toastTilesSearch.forEach(toasterSearch => {
                    let toasterName = toasterSearch.getAttribute("data-toastername");
                    if (!toasterName.toLowerCase().includes(inpSearchToaster.value.trim().toLowerCase())) {
                        toasterSearch.classList.add("hidden");
                    }
                });

                var toastersNoHidden =  document.querySelectorAll(".toastTile__item.from-search:not(.hidden)");
                gSearchToasterCount = toastersNoHidden ? toastersNoHidden.length : 0;
                tileResultElem.textContent = gSearchToasterCount;

                if (gSearchToasterCount === 0) {
                    emptyResultToasterElem.classList.add("show");
                }
                else if (emptyResultToasterElem.classList.contains("show")) {
                    emptyResultToasterElem.classList.remove("show");
                }
            }
        },
        emptyCallback: () => {
            var emptyZone = document.querySelector(".empty__zone");
            var emptyZoneSearch = document.querySelector(".toastTile__empty-from-search");

            if (emptyZone) { emptyZone.classList.add("show"); }
            if (emptyZoneSearch) { emptyZoneSearch.classList.add("show"); }
        }
    });
}

function handlePropagateMyToasters() {
    var blockToasters = document.querySelector(".section-toasters:not(.last-toasters)");
    var blockLastToasters = document.querySelector(".section-toasters.last-toasters");

    handleDisplaySearchToasters();

    // propagate to my toasters
    if (blockToasters) handleRenderToasters(blockToasters);
    if (blockLastToasters) {
        handleDisplayToasters()
    };
}

function closeSearchScreen() {
    searchScreen.classList.remove("show");
    document.documentElement.style.overflow = "auto";

    timeoutSearchToaster = setTimeout(() => {
        searchScreen.style.display = "none";
    }, 300);

    inpSearchToaster.value = "";
}

/* handle action of header account menu */
document.querySelectorAll(".header-account__menu-item:not(.menu-item__link)").forEach(headerItem => {
    headerItem.addEventListener("click", function (e) {
        e.preventDefault();

        let action = this.getAttribute("data-action");

        switch (action) {
            case "signout":
                handleSignout();
                break;
            default:
        }
    });
});

function handleSignout() {
    signout().then(cb => {
        if (cb && cb.success) {
            window.location = "/login";
        }
        else if (cb && !cb.success) {
            ALERT_MOD.call({
                title: "Information",
                text: cb.code + " : " + cb.message
            });
        }
    });
}
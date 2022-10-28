import "local://includes/routes/user.js";
import "local://includes/sidemenu.js";
import "local://includes/alert.js";

var USER = null;

document.addEventListener("DOMContentLoaded", function () {
    getUser().then(cb => {
        if (cb && cb.success) {
            USER = cb.user;
            
            new Router();

            if (USER && showContent(USER)) {
                setUserContent(USER);
                initSidemenu();
            }
        }
        else {
            new Router();
        }
    });
});

/**
 * 
 * @param {Object} USER 
 */
function showContent(USER) {
    let isSuccess = false;

    if (USER) {
        var temp = document.getElementsByTagName("template")[0];
        let pageAccess = JSON.parse(temp.getAttribute("data-access"));
        let userType = getUserType(USER);

        if (pageAccess.includes(userType) || pageAccess.includes("all")) {
            var cloned = temp.content.cloneNode(true);
            document.body.appendChild(cloned);

            isSuccess = true;
        }
    }

    return isSuccess;
}

/**
 * 
 * @param {Object} USER 
 */
function setUserContent(USER) {
    var usernameHeaderElem = document.querySelector(".header-account__user");
    var headerAccountPlanElem = document.querySelector(".header-account__plan");
    var sectionStartSubElem = document.querySelector(".section-start-subscription");
    var headerProfileElem = document.querySelector(".header-account__profile-img");

    usernameHeaderElem.textContent = USER.username;

    if (headerAccountPlanElem && USER.active_billing) {
        headerAccountPlanElem.style.display = "block";

        setTimeout(() => {
            headerAccountPlanElem.classList.add("show");
        }, 100);
    }

    if (sectionStartSubElem && USER.active_billing) {
        sectionStartSubElem.remove();
    }

    if (USER.picture_link) {
        headerProfileElem.src = USER.picture_link;
    }
    else {
        headerProfileElem.src = "/assets/images/no-profile.png";
    }
}
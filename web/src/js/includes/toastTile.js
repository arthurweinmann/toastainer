function setToastTileEvents(blockToasters, options) {
    var toastersMoreElems = blockToasters.querySelectorAll(".toastTile__more");
    var toastersTileElems = blockToasters.querySelectorAll(".toastTile__item");
    var toastersBtnEditElems = blockToasters.querySelectorAll(".btn-edit-toaster");
    var toastersBtnDeleteElems = blockToasters.querySelectorAll(".btn-delete-toaster");

    toastersBtnEditElems.forEach(elem => {
        elem.addEventListener("click", function () {
            let toasterId = this.getAttribute("data-toasterId");
            window.location.href = "/edit-toaster?id=" + toasterId;
        });
    });

    toastersBtnDeleteElems.forEach(elem => {
        elem.addEventListener("click", function () {
            let toasterId = this.getAttribute("data-toasterId");

            ALERT_MOD.call({
                title: "Confirmation",
                text: "Are you sure to delete this Toaster ?",
                buttons: [
                    {
                        text: "CANCEL",
                        onClick: function () {
                            ALERT_MOD.close();
                        }
                    },
                    {
                        text: "DELETE",
                        onClick: function () {
                            ALERT_MOD.close();
                            deleteToaster(toasterId).then(cb => {
                                if (cb && cb.success) {
                                    if (options.deleteCallback) {
                                        options.deleteCallback();
                                    }
                                    else {
                                        handleRenderToasters(blockToasters, options);
                                    }
                                }
                                else if (cb && !cb.success) {
                                    ALERT_MOD.call({
                                        title: "Information",
                                        text: cb.code + " : " + cb.message
                                    });
                                }
                            });
                        }
                    },
                ]
            });
        });
    });

    toastersMoreElems.forEach(elem => {
        elem.addEventListener("click", function () {
            var toasterDetailsMenu = this.parentNode.parentNode.querySelector(".toastTile__menu");

            if (toasterDetailsMenu.classList.contains("show-menu")) {
                this.classList.remove("show-menu");
                toasterDetailsMenu.classList.remove("show-menu");
            }
            else {
                this.classList.add("show-menu");
                toasterDetailsMenu.classList.add("show-menu")
            }
        });
    });

    toastersTileElems.forEach(toastTile => {
        const delta = 6;
        let startX;
        let startY;

        toastTile.addEventListener("mousedown", function (event) {
            startX = event.pageX;
            startY = event.pageY;
        });

        toastTile.addEventListener("mouseup", function (event) {
            const diffX = Math.abs(event.pageX - startX);
            const diffY = Math.abs(event.pageY - startY);

            let toastTileMore = this.querySelector(".toastTile__more");

            if (diffX < delta && diffY < delta) {       // is click
                let isMore = toastTileMore.contains(event.target);

                if (!isMore && event.target.className !== "btn-edit-toaster" && event.target.className !== "btn-delete-toaster") {  // redirect to Toaster page
                    let toasterId = this.getAttribute("data-toasterId");
                    window.location.href = "/toaster?id=" + toasterId;
                }
            }
            else {                                      // is drag
            }
        });
    });
}

function handleRenderToasters(blockToasters, options) {
    var countElem = blockToasters.querySelector(".tile__count");
    var listToastersElem = blockToasters.querySelector(".toastTile__list");

    if (options && options.sliderListClass) {
        listToastersElem = blockToasters.querySelector(options.sliderListClass);
        blockToasters = listToastersElem;
    }

    listToasters().then(cb => {
        if (cb && cb.success) {
            var toastersArray = cb.toasters ? cb.toasters.slice() : [];

            if (countElem) { countElem.textContent = toastersArray.length; }
            if (options && options.emptyCallback && toastersArray.length == 0) { options.emptyCallback(); }
            if (options && options.filterToasters) {
                toastersArray = options.filterToasters(toastersArray);
            }
            if (options && options.isFromSearch) {
                gListToastersForSearch = toastersArray;
            }

            renderToasters(listToastersElem, toastersArray, options);
            setToastTileEvents(blockToasters, options);

            if (options && options.endCallback) {
                options.endCallback();
            }
        }
        else if (cb && !cb.success) {
            ALERT_MOD.call({
                title: "Information",
                text: cb.code + " : " + cb.message
            });
        }
    });
}

function renderToasters(listToastersElem, toasters, options) {
    var string = "";
    var toastersCloned = toasters.slice();
    var isFirst = true;

    if (listToastersElem) {
        listToastersElem.innerHTML = "";
    }

    if (options && options.isItemSlider) {
        for (let i = toastersCloned.length; i < 5; i++) {
            if (isFirst) {
                toastersCloned.splice(0, 0, null);
                isFirst = false;
            }
            else {
                toastersCloned.push(null);
            }
        }
    }

    for (let i = 0; i < toastersCloned.length; i++) {
        let toaster = toastersCloned[i];
        string += getToasterTileHTML(toaster, options);
    }

    listToastersElem.innerHTML = string;
}

function getToasterTileHTML(toaster, options) {
    let isItemSlider = options && options.isItemSlider;
    let isMostCalled = options && options.isMostCalled;
    let isMostRecent = options && options.isMostRecent;
    let isFromSearch = options && options.isFromSearch;
    let hasUserPicture = USER.picture_link ? true : false;

    if (toaster !== null) {
        let hasToasterPicture = toaster.picture_link ? true : false;

        return `<li class="toastTile__item` +
            (isItemSlider ? " item-slider" : "") +
            (isMostCalled ? " most-called" : "") +
            (isMostRecent ? " most-recent" : "") +
            (isFromSearch ? " from-search" : "") +
            `" data-toasterId="${toaster.id}" data-toastername="${toaster.name ? toaster.name : getShortDate(new Date(toaster.created * 1000))}">
            <div class="toastTile">
                <div class="toastTile__wrapper">
                    <div class="toastTile__toasty" `+ (hasToasterPicture ? `style="background-image: url('${toaster.picture_link}'); background-color: transparent; border: 1px solid rgba(255,255,255,0.1);"` : "") + `>
                        ` + (!hasToasterPicture ? `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 576 512">
                            <path d="M288 0C108 0 0 93.4 0 169.14 0 199.44 24.24 224 64 224v256c0 17.67 16.12 32 36 32h376c19.88 0 36-14.33 36-32V224c39.76 0 64-24.56 64-54.86C576 93.4 468 0 288 0z" />
                        </svg>`: "") + `   
                    </div>
                    <div class="toastFile__main">
                        <p class="toastTile__name">${toaster.name ? toaster.name : getShortDate(new Date(toaster.created * 1000))}</p>
                        <p class="toastTile__infos">Toasted by 
                            <span class="toastTile__profile" style="background-image: url('`+ (hasUserPicture ? USER.picture_link : "/assets/images/no-profile.png") + `')"></span> 
                            <span class="toastTile__infos-author">${USER.username}</span>
                        </p>
                    </div>
                    <div class="toastTile__more">
                        <span class="icon-chevron-down"></span>
                    </div>
                </div>
                <div class="toastTile__menu">
                    <p class="toastTile__details">Details</p>
                    <ul>
                        <li class="toastTile__item-detail">
                            <span class="toastTile__label-menu">ID</span>
                            <span class="toastTile__detail-val">${toaster.id}</span>
                        </li>
                        <li class="toastTile__item-detail">
                        <span class="toastTile__label-menu">UP</span>
                        <span class="toastTile__detail-val">Updated ${timeSince(new Date(toaster.last_modified))}</span>
                    </li>
                        <li class="toastTile__item-btn">
                            <button class="btn-edit-toaster" data-toasterId="${toaster.id}">Edit</button>
                        </li>
                        <li class="toastTile__item-btn">
                            <button class="btn-delete-toaster" data-toasterId="${toaster.id}">Delete</button>
                        </li>
                    </ul>
                </div>
            </div>
        </li>`;
    }
    else {
        return `<li class="toastTile__item item-slider toastTile__empty-slider disabled">
            <div class="toastTile" style="opacity: 0;">
                <div class="toastTile__wrapper">
                    <div class="toastTile__toasty">
                        <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 576 512">
                            <path d="M288 0C108 0 0 93.4 0 169.14 0 199.44 24.24 224 64 224v256c0 17.67 16.12 32 36 32h376c19.88 0 36-14.33 36-32V224c39.76 0 64-24.56 64-54.86C576 93.4 468 0 288 0z" />
                        </svg>
                    </div>
                    <div class="toastFile__main">
                        <p class="toastTile__name">Empty</p>
                        <p class="toastTile__infos">Toasted by 
                            <span class="toastTile__profile" style="background-image: url('/assets/images/no-profile.png')"></span> 
                            <span class="toastTile__infos-author">empty</span>
                        </p>
                    </div>
                    <div class="toastTile__more">
                        <span class="icon-chevron-down"></span>
                    </div>
                </div>
            </div>
        </li>`;
    }
}

// TODO when toastfront can includes/utils from includes/file.js
function getShortDate(date) {
    return date.toLocaleDateString('fr-FR');
}

function timeSince(date) {
    var seconds = Math.floor(((new Date().getTime() / 1000) - date))

    var interval = seconds / 31536000;

    if (interval >= 1) {
        let intervalFloor = Math.floor(interval);
        return intervalFloor === 1 ? "Last year " : intervalFloor + " years ago";
    }
    interval = seconds / 2592000;
    if (interval >= 1) {
        let intervalFloor = Math.floor(interval);
        return intervalFloor === 1 ? "Last month" : intervalFloor + " months ago";
    }
    interval = seconds / 86400;
    if (interval >= 1) {
        let intervalFloor = Math.floor(interval);
        return intervalFloor === 1 ? "Yesterday" : intervalFloor + " days ago";
    }
    interval = seconds / 3600;
    if (interval >= 1) {
        let intervalFloor = Math.floor(interval);
        return intervalFloor === 1 ? "1 hour ago" : intervalFloor + " hours ago";
    }
    interval = seconds / 60;
    if (interval >= 1) {
        let intervalFloor = Math.floor(interval);
        return intervalFloor === 1 ? "1 minute ago" : intervalFloor + " minutes ago";
    }
    return Math.floor(seconds) + " seconds ago";
}
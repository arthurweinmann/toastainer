function ModalLinkToaster() {
    this.MODAL_CLASS = "modal-link-toaster";

    this.defaultsCall = {
        onInit: null,
        title: null,
        text: null,
        linkedToaster: "",
        toasters: [],
        buttons: [{
            text: "OK",
            onClick: null
        }]
    }

    this.init();
}

ModalLinkToaster.prototype.constructor = ModalLinkToaster;

ModalLinkToaster.prototype.init = function () { }

ModalLinkToaster.prototype.close = function () {
    let modalClass = document.querySelector("." + this.MODAL_CLASS);

    if (modalClass) {
        modalClass.remove();
    }
}

ModalLinkToaster.prototype.call = function (options) {
    var defaultObj = JSON.parse(JSON.stringify(this.defaultsCall));

    var options = Object.assign(defaultObj, options);

    if (defaultObj.buttons.length > 2) { alert("Max number of buttons: 2"); return false; }

    MODAL_LINK_TOASTER.close();

    this.templateToaster = templateToaster(options);

    document.getElementById("wrapper-modal-link-toaster").innerHTML = templateToaster(options);

    document.getElementById("modal-inp-search-toaster").addEventListener("keyup", function (e) {
        let val = e.target.value;

        let toastersFiltered = options.toasters.filter(toaster => {
            return toaster.name.toLowerCase().includes(val.trim().toLowerCase());
        });

        document.querySelectorAll(".toaster__item").forEach(toasterItemElem => {
            toasterItemElem.style.display = "flex";
        });

        document.querySelectorAll(".toaster__item").forEach(toasterItemElem => {
            let toasterId = toasterItemElem.getAttribute("data-toaster-id");
            if (toastersFiltered.findIndex(toaster => toaster.id === toasterId) < 0) {
                toasterItemElem.style.display = "none";
            }
        });
    });

    createButtonEventToaster(options);
    handleInitToaster(options);
}

ModalLinkToaster.prototype.destroy = function () {

}

// called when the item was added in DOM
function handleInitToaster(option) {
    if (option.onInit) {
        option.onInit.call();
    }
}

// create buttons events 
function createButtonEventToaster(option) {
    document.querySelectorAll(".modal-link-toaster__button").forEach(alertBtn => {
        alertBtn.addEventListener("click", function () {
            let id = this.getAttribute("data-id");
            let obj = option.buttons[id];

            //check if there is a custom function onClick or not
            if (obj.onClick) {
                obj.onClick.call();
            }
            else {
                MODAL_LINK_TOASTER.close();
            }

        });
    });
}

// alert templateToaster 
function templateToaster(options) {
    let buttons = options.buttons;
    let btnHTML = ``;

    for (let i = 0; i < buttons.length; i++) {
        let btn = buttons[i];
        btnHTML += `<span class="modal-link-toaster__button" data-id="` + i + `">` + btn.text + `</span>`
    }

    let toastersHTML = ``;

    for (let i = 0; i < options.toasters.length; i++) {
        let toaster = options.toasters[i];

        toastersHTML += `<li class="toaster__item" data-toaster-id="` + toaster.id + `">
            <label class="checkbox">
                <input type="radio" name="toaster" value="`+ toaster.id + `" ` + (options.linkedToaster === toaster.id ? "checked" : "") + `>
                `+ toaster.name + `
            </label>
        </li>`;
    }

    var templateToaster = `<div class="modal-link-toaster">
            <div class="modal-link-toaster__block"></div>
                <table class="modal-link-toaster__box">
                    <tr>
                        <td class="modal-link-toaster__content">
                            <div class="modal-link-toaster__wrapper">
                                <div class="modal-link-toaster__inner">
                                    `+ (options.title ? `<div class="modal-link-toaster__title">` + options.title + `</div>` : "") + `
                                    `+ (options.text ? `<div class="modal-link-toaster__text">` + options.text + `</div>` : "") + `
                                     
                                    <div class="modal-cont-search">
                                        <input id="modal-inp-search-toaster" placeholder="Search Toaster" />
                                        <div class="modal-cont-icon-search">
                                            <span class="icon-search"></span>
                                        </div>
                                    </div>

                                    <ul class="toaster__list">
                                       ` + toastersHTML + `
                                    </ul>
                                </div>
                                <div class="modal-link-toaster__buttons modal-link-toaster__buttons--` + buttons.length + `">
                                    `+ btnHTML + `
                                </div>
                            </div>
                        </td>                        
                    </tr>
                </table>
            </div>
        </div>`;

    return templateToaster;
}

var MODAL_LINK_TOASTER = new ModalLinkToaster();

function ModalLinkCustomDomain() {
    this.MODAL_CLASS = "modal-link-custom-domain";

    this.defaultsCall = {
        onInit: null,
        title: null,
        text: null,
        toasterId: "",
        customDomains: [],
        buttons: [{
            text: "OK",
            onClick: null
        }]
    }

    this.init();
}

ModalLinkCustomDomain.prototype.constructor = ModalLinkCustomDomain;

ModalLinkCustomDomain.prototype.init = function () { }

ModalLinkCustomDomain.prototype.close = function () {
    let modalClass = document.querySelector("." + this.MODAL_CLASS);

    if (modalClass) {
        modalClass.remove();
    }
}

ModalLinkCustomDomain.prototype.call = function (options) {
    var defaultObj = JSON.parse(JSON.stringify(this.defaultsCall));

    var options = Object.assign(defaultObj, options);

    if (defaultObj.buttons.length > 2) { alert("Max number of buttons: 2"); return false; }

    MODAL_LINK_CUSTOM_DOM.close();

    this.templateCustomDoms = templateCustomDoms(options);

    document.getElementById("wrapper-modal-link-custom-domain").innerHTML = templateCustomDoms(options);

    var customDomainsMoreElems = document.querySelectorAll(".custom-domain__more");
    customDomainsMoreElems.forEach(elem => {
        elem.addEventListener("click", function () {
            if (this.classList.contains("show-menu")) {
                this.classList.remove("show-menu");

                var domainList = this.parentNode.parentNode.querySelector(".domain__list");

                this.parentNode.parentNode.style.height = "min-content";
                domainList.style.height = "0";
                domainList.style.paddingTop = "0";
            }
            else {
                this.classList.add("show-menu");

                var domainList = this.parentNode.parentNode.querySelector(".domain__list");

                this.parentNode.parentNode.style.height = "initial";
                domainList.style.height = "initial";
                domainList.style.paddingTop = "10px";
            }
        });
    });

    createButtonEventCustomDoms(options);
    handleInitCustomDoms(options);
}

ModalLinkCustomDomain.prototype.destroy = function () {

}

// called when the item was added in DOM
function handleInitCustomDoms(option) {
    if (option.onInit) {
        option.onInit.call();
    }
}

// create buttons events 
function createButtonEventCustomDoms(option) {
    document.querySelectorAll(".modal-link-custom-domain__button").forEach(alertBtn => {
        alertBtn.addEventListener("click", function () {
            let id = this.getAttribute("data-id");
            let obj = option.buttons[id];

            //check if there is a custom function onClick or not
            if (obj.onClick) {
                obj.onClick.call();
            }
            else {
                MODAL_LINK_CUSTOM_DOM.close();
            }

        });
    });
}

// alert templateCustomDoms 
function templateCustomDoms(options) {
    let buttons = options.buttons;
    let btnHTML = ``;

    for (let i = 0; i < buttons.length; i++) {
        let btn = buttons[i];
        btnHTML += `<span class="modal-link-custom-domain__button" data-id="` + i + `">` + btn.text + `</span>`
    }

    let customDomainsHTML = ``;

    for (let i = 0; i < options.customDomains.length; i++) {
        let cDomain = options.customDomains[i];

        let subDomainsHTML = ``;

        for (let j = 0; j < cDomain.subdomains.length; j++) {
            let subDom = cDomain.subdomains[j];

            subDomainsHTML += `<li class="domain__item">
                <label class="checkbox">
                    <input type="checkbox" value="`+ subDom + `" data-custom-domain="` + cDomain.id + `" ` + (cDomain.linked_toasters && cDomain.linked_toasters[subDom] === options.toasterId ? "checked" : "") + `>
                    `+ subDom + `
                </label>
            </li>`;
        }

        customDomainsHTML += `<li class="custom-domain__item">
                <div class="custom-domain__tab"><span class="custom-domain__name">`+ cDomain.root_domain + `</span><div class="custom-domain__more show-menu"><span class="icon-chevron-down"></div></span></div>
                <ul class="domain__list">
                    `+ subDomainsHTML + `
                </ul>
        </li>`;
    }

    var templateCustomDoms = `<div class="modal-link-custom-domain">
            <div class="modal-link-custom-domain__block"></div>
                <table class="modal-link-custom-domain__box">
                    <tr>
                        <td class="modal-link-custom-domain__content">
                            <div class="modal-link-custom-domain__wrapper">
                                <div class="modal-link-custom-domain__inner">
                                    `+ (options.title ? `<div class="modal-link-custom-domain__title">` + options.title + `</div>` : "") + `
                                    `+ (options.text ? `<div class="modal-link-custom-domain__text">` + options.text + `</div>` : "") + `
                                     
                                    <ul class="custom-domain__list">
                                       ` + customDomainsHTML + `
                                    </ul>
                                </div>
                                <div class="modal-link-custom-domain__buttons modal-link-custom-domain__buttons--` + buttons.length + `">
                                    `+ btnHTML + `
                                </div>
                            </div>
                        </td>                        
                        
                    </tr>
                </table>
            </div>
        </div>`;

    return templateCustomDoms;
}

var MODAL_LINK_CUSTOM_DOM = new ModalLinkCustomDomain();

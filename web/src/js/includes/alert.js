function Alert() {
    this.MODAL_CLASS = "modal-alert";

    this.defaultsCall = {
        onInit: null,
        title: null,
        text: null,
        loadingText: null,
        isLoading: false,
        withCheckmark: false,
        buttons: [{
            text: "OK",
            onClick: null
        }]
    }

    this.init();
}

Alert.prototype.constructor = Alert;

Alert.prototype.init = function () { }

Alert.prototype.close = function () {
    let modalClass = document.querySelector("." + this.MODAL_CLASS);

    if (modalClass) {
        modalClass.remove();
    }
}

Alert.prototype.call = function (options) {
    var defaultObj = JSON.parse(JSON.stringify(this.defaultsCall));

    var options = Object.assign(defaultObj, options);

    if (defaultObj.buttons.length > 2) { alert("Max number of buttons: 2"); return false; }

    ALERT_MOD.close();

    this.templateAlert = templateAlert(options);

    document.getElementById("wrapper-modal").innerHTML = templateAlert(options);

    createButtonEventAlert(options);
    handleInitAlert(options);
}

Alert.prototype.destroy = function () {

}

// called when the item was added in DOM
function handleInitAlert(option) {
    if (option.onInit) {
        option.onInit.call();
    }
}

// create buttons events 
function createButtonEventAlert(option) {
    document.querySelectorAll(".modal-alert__button").forEach(alertBtn => {
        alertBtn.addEventListener("click", function () {
            let id = this.getAttribute("data-id");
            let obj = option.buttons[id];

            //check if there is a custom function onClick or not
            if (obj.onClick) {
                obj.onClick.call();
            }
            else {
                ALERT_MOD.close();
            }

        });
    });
}

// alert templateAlert 
function templateAlert(options) {
    let buttons = options.buttons;
    let btnHTML = ``;

    for (let i = 0; i < buttons.length; i++) {
        let btn = buttons[i];
        btnHTML += `<span class="modal-alert__button" data-id="` + i + `">` + btn.text + `</span>`
    }

    var templateAlert = `<div class="modal-alert">
            <div class="modal-alert__block"></div>
                <table class="modal-alert__box">
                    <tr>
                        <td class="modal-alert__content">
                            <div class="modal-alert__wrapper">
                                <div class="modal-alert__inner">
                                    `+ (options.title ? `<div class="modal-alert__title">` + options.title + `</div>` : "") + `
                                    `+ (options.isLoading ? `<div class="stage-toaster-loading">
                                        <div class="breads">
                                        <div class="bread bread-left"></div>
                                        <div class="bread bread-right"></div>
                                        </div>
                                        <div class="toaster">
                                        <div class="eyes">
                                            <div class="eye eye-left"></div>
                                            <div class="eye eye-right"></div>
                                        </div>
                                        <div class="eyes eyes-second">
                                            <div class="eye-second eye-left"></div>
                                            <div class="eye-second eye-right"></div>
                                        </div>
                                        <div class="mouth"></div>
                                        <div class="toaster-white"></div>
                                        </div>
                                        <div class="toaster-bottom">
                                        <div class="toaster-bottom-white"></div>
                                        </div>
                                        <div class="bottom-line">
                                        <div class="line"></div>
                                        <div class="line"></div>
                                        <div class="line"></div>
                                        </div>
                                    </div>`: "") + `
                                    `+ (options.loadingText ? `<div class="modal-alert__loading">` + options.loadingText + `</div>` : "") + `
                                    `+ (options.text ? `<div class="modal-alert__text">` + options.text + `</div>` : "") + `
                                    `+ (options.withCheckmark ? `<svg class="checkmark" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 52 52"><circle class="checkmark__circle" cx="26" cy="26" r="25" fill="none"/><path class="checkmark__check" fill="none" d="M14.1 27.2l7.1 7.2 16.7-16.8"/></svg>` : "") + `
                                </div>
                                `+ (!options.isLoading ? `<div class="modal-alert__buttons modal-alert__buttons--` + buttons.length + `">`+ btnHTML + ` </div>`: "") + `
                            </div>
                        </td>                        
                        
                    </tr>
                </table>
            </div>
        </div>`;

    return templateAlert;
}

var ALERT_MOD = new Alert();

// exemples button
// document.querySelector(".js-button-1").addEventListener("click", function () {
//     ALERT_MOD.call({
//         title: "Custom title",
//         text: "A custom alert with confirm and cancel button",
//         buttons: [
//             {
//                 text: "CONFIRM",
//                 onClick: function () {
//                     ALERT_MOD.call({ title: "Hey man!", text: "Thank u for checking this!" });
//                 }
//             },
//             {
//                 text: "CANCEL",
//                 onClick: function () {
//                     ALERT_MOD.close();
//                 }
//             }
//         ]
//     });
//     return false;
// });

// document.querySelector(".js-button-2").addEventListener("click", function () {
//     ALERT_MOD.call({
//         onInit: function () { },
//         title: "Custom title",
//         text: null,
//         buttons: [{
//             text: "OK",
//             onClick: function () {
//                 ALERT_MOD.close();
//             }
//         }]
//     });
//     return false;
// });

// document.querySelector(".js-button-3").addEventListener("click", function () {
//     ALERT_MOD.call({ title: "That"s all folks!", text: "Custom text comes here" });
//     return false;
// });

// document.querySelector(".js-button-4").addEventListener("click", function () {
//     ALERT_MOD.call({ text: "Custom text comes here" });
//     return false;
// });

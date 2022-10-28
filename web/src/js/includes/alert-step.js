function AlertStep() {
    this.MODAL_CLASS = "modal-alert-step";

    this.defaultsCall = {
        onInit: null,
        title: null,
        text: null,
        step: 0,
        nbSteps: 0,
        buttons: [{
            text: "OK",
            onClick: null
        }]
    }

    this.init();
}

AlertStep.prototype.constructor = AlertStep;

AlertStep.prototype.init = function () { }

AlertStep.prototype.close = function () {
    let modalClass = document.querySelector("." + this.MODAL_CLASS);

    if (modalClass) {
        modalClass.remove();
    }
}

AlertStep.prototype.call = function (options) {
    var defaultObj = JSON.parse(JSON.stringify(this.defaultsCall));

    var options = Object.assign(defaultObj, options);

    if (defaultObj.buttons.length > 2) { alert("Max number of buttons: 2"); return false; }

    ALERT_STEP_MOD.close();

    this.templateAlertStep = templateAlertStep(options);

    document.getElementById("wrapper-modal").innerHTML = templateAlertStep(options);

    createButtonEventAlertStep(options);
    handleInitAlertStep(options);
}

AlertStep.prototype.destroy = function () {

}

// called when the item was added in DOM
function handleInitAlertStep(option) {
    if (option.onInit) {
        option.onInit.call();
    }
}

// create buttons events 
function createButtonEventAlertStep(option) {
    document.querySelectorAll(".modal-alert-step__button").forEach(alertBtn => {
        alertBtn.addEventListener("click", function () {
            let id = this.getAttribute("data-id");
            let obj = option.buttons[id];

            //check if there is a custom function onClick or not
            if (obj.onClick) {
                obj.onClick.call();
            }
            else {
                ALERT_STEP_MOD.close();
            }

        });
    });
}

// alert templateAlertStep 
function templateAlertStep(options) {
    let buttons = options.buttons;
    let btnHTML = ``;

    for (let i = 0; i < buttons.length; i++) {
        let btn = buttons[i];
        btnHTML += `<span class="modal-alert-step__button" data-id="` + i + `">` + btn.text + `</span>`
    }

    var templateAlertStep = `<div class="modal-alert-step">
            <div class="modal-alert-step__block"></div>
                <table class="modal-alert-step__box">
                    <tr>
                        <td class="modal-alert-step__content">
                            <div class="modal-alert-step__wrapper">
                                <div class="modal-alert-step__inner">
                                    <div class="modal-alert-step__step">
                                        <span>`+ options.step + `/` + options.nbSteps + `</span>
                                    </div>
                                    `+ (options.title ? `<div class="modal-alert-step__title">` + options.title + `</div>` : "") + `
                                    `+ (options.text ? `<div class="modal-alert-step__text">` + options.text + `</div>` : "") + `
                                    <div class="modal-alert-step__warn">Please don't leave until linking process ended.</div>
                                </div>
                                <div class="modal-alert-step__buttons modal-alert-step__buttons--` + buttons.length + `">
                                    `+ btnHTML + `
                                </div>
                            </div>
                        </td>                        
                    </tr>
                </table>
            </div>
        </div>`;

    return templateAlertStep;
}

var ALERT_STEP_MOD = new AlertStep();

import "local://includes/routes/user.js";
import "local://includes/loading-utils.js";
import "local://includes/utils/string.js";

document.addEventListener("DOMContentLoaded", () => {
    var form = document.querySelector(".form");
    var formErr = document.querySelector(".form-err");
    var errTextElem = document.querySelector(".err__text");

    form.addEventListener("submit", function (e) {
        e.preventDefault();

        let formElements = this.elements;
        let data = {
            email: formElements["email"].value.toLowerCase().trim(),
            password: formElements["password"].value.trim()
        };

        ALERT_MOD.call({
            isLoading: true,
            loadingText: "Loading..."
        });

        let promise = waitAtLeast(800, signin(data.email, data.password));
        promise.then((cb) => {
            if (cb && cb.success) {
                window.location = "/";
            }
            else if (cb && !cb.success) {
                formErr.classList.add("show");
                errTextElem.textContent = capitalizeFirstLetter(cb.message) + ".";
                ALERT_MOD.close();
            }
        });
    });
});
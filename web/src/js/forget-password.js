import "local://includes/routes/user.js";
import "local://includes/loading-utils.js";

document.addEventListener("DOMContentLoaded", () => {
    var form = document.querySelector(".form");
    var formErr = document.querySelector(".form-err");

    form.addEventListener("submit", function (e) {
        e.preventDefault();

        ALERT_MOD.call({
            isLoading: true,
            loadingText: "Loading..."
        });

        let formElements = this.elements;
        let email = formElements["email"].value.toLowerCase().trim();
        let promise = waitAtLeast(800, forgetPassword(email));

        promise.then((cb) => {
            if (cb && cb.success) {
                ALERT_MOD.close();
                document.querySelector(".form").remove();
                document.querySelector(".title-h1").textContent = "Check your email and spam folder";
                document.querySelector(".infos__text").textContent = "We have sent a password reset link. If the email does not show up in a few minutes, check your spam folder or try again.";
            }
            else if (cb && !cb.success) {
                var formErrTextElem = document.querySelector(".form-err__text");
                ALERT_MOD.close();
                formErr.classList.add("show");

                if (cb.code === "notFound") {
                    formErrTextElem.textContent = "No account found for this email.";
                }
                else {
                    formErrTextElem.textContent = "An error occured. Please, try again or contact us.";
                }
            }
        });
    });
});
import "local://includes/routes/user.js";
import "local://includes/loading-utils.js";

var tokenFromURL = new URLSearchParams(window.location.search).get("token");


document.addEventListener("DOMContentLoaded", () => {
    if (tokenFromURL) {
        var form = document.querySelector(".form");
        var formErr = document.querySelector(".form-err");

        form.addEventListener("submit", function (e) {
            e.preventDefault();

            ALERT_MOD.call({
                isLoading: true,
                loadingText: "Loading..."
            });

            var formElements = this.elements;
            let data = {
                token: tokenFromURL,
                password: formElements["password"].value.trim()
            };

            let promise = waitAtLeast(800, resetPassword(data.token, data.password));              // reset password

            promise.then((cb) => {                                  
                if (cb && cb.success) {
                    ALERT_MOD.close();
                    document.querySelector(".form").remove();
                    document.querySelector(".title-h1").textContent = "Password changed !";
                    document.querySelector(".cont-success").innerHTML = `<p class="infos__text">Your password was successfully reset.</p>
                    <a class="btn" href="/login">Sign in</a>`;
                }
                else if (cb && !cb.success) {
                    ALERT_MOD.close();
                    formErr.classList.add("show");
                }
            });
        });
    }
    else {
        ALERT_MOD.call({
            title: "Information",
            text: "Token to reset password is missing."
        });
    }

});
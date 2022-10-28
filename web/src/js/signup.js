import "local://includes/routes/user.js";
import "local://includes/loading-utils.js";
import "local://includes/utils/string.js";

document.addEventListener("DOMContentLoaded", () => {
    const regexMail = new RegExp(/^\w+([\.-]?\w+)*@\w+([\.-]?\w+)*(\.\w{2,3})+$/);     // regex for email
    const regex1 = new RegExp(/.{8,}$/);                                               // at least 8 characters
    const regex2 = new RegExp(/(?=.*[0-9])/);                                          // at least one number
    const regex3 = new RegExp(/[ `!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]/);             // at least one symbol

    var form = document.querySelector(".form");
    var formErr = document.querySelector(".form-err");
    var errTextElem = document.querySelector(".err__text");
    var inpEmailElem = document.getElementById("inp-email");
    var inpUsernameElem = document.getElementById("inp-username");
    var inpPasswordElem = document.getElementById("inp-password");
    var validationItemsElem = document.querySelectorAll(".pwd-validation__item");
    var btnSignup = document.getElementById("btn-signup");
    let isEmailValid = false;
    let isUsernameValid = false;
    let isPasswordValid = false;

    inpEmailElem.addEventListener("change", function (e) {
        let isValidRegexMail = regexMail.test(e.target.value);

        if (isValidRegexMail) {
            isEmailValid = true;
        }
        else {
            isEmailValid = false;
        }
        checkSignupButton();
    });

    inpUsernameElem.addEventListener("change", function (e) {
        if (e.target.value.length > 0) {
            isUsernameValid = true;
        }
        else {
            isUsernameValid = false;
        }
        checkSignupButton();
    });

    inpPasswordElem.addEventListener("change", function (e) {
        let pwdValue = e.target.value;
        let isValidRegex1 = regex1.test(pwdValue);
        let isValidRegex2 = regex2.test(pwdValue);
        let isValidRegex3 = regex3.test(pwdValue);

        isValidRegex1 ? validationItemsElem[0].classList.add("is-valid") : validationItemsElem[0].classList.remove("is-valid");
        isValidRegex2 ? validationItemsElem[1].classList.add("is-valid") : validationItemsElem[1].classList.remove("is-valid");
        isValidRegex3 ? validationItemsElem[2].classList.add("is-valid") : validationItemsElem[2].classList.remove("is-valid");

        if (isValidRegex1 && isValidRegex2 && isValidRegex3) {
            isPasswordValid = true;
        }
        else {
            isPasswordValid = false;
        }

        checkSignupButton();
    });

    form.addEventListener("submit", function (e) {
        e.preventDefault();

        if (isEmailValid && isUsernameValid && isPasswordValid) {
            formErr.classList.remove("show");

            let formElements = this.elements;
            let data = {
                email: formElements["email"].value.toLowerCase().trim(),
                username: formElements["username"].value.toLowerCase().trim(),
                password: formElements["password"].value.trim()
            };

            ALERT_MOD.call({
                isLoading: true,
                loadingText: "Loading..."
            });

            let promise = waitAtLeast(800, signup(data.email, data.username, data.password));
            promise.then((cb) => {
                if (cb && cb.success) {
                    signin(data.email, data.password).then(cb => {
                        if (cb && cb.success) {
                            ALERT_MOD.call({
                                title: "Success",
                                text: "Your account was successfully created.",
                                withCheckmark: true,
                                buttons: [
                                    {
                                        text: "Go to the platform",
                                        onClick: function () {
                                            ALERT_MOD.close();
                                            window.location = "/";
                                        }
                                    },
                                ]
                            });
                        }
                        else {
                            ALERT_MOD.call({
                                title: "Success",
                                text: "Your account was successfully created.",
                                withCheckmark: true,
                                buttons: [
                                    {
                                        text: "Let's sign in",
                                        onClick: function () {
                                            ALERT_MOD.close();
                                            window.location = "/login";
                                        }
                                    },
                                ]
                            });
                        }
                    })
                }
                else if (cb && !cb.success) {
                    formErr.classList.add("show");
                    errTextElem.textContent = capitalizeFirstLetter(cb.message) + ".";
                    ALERT_MOD.close();
                }
            });
        }
    });

    function checkSignupButton() {
        if (isEmailValid && isUsernameValid && isPasswordValid) {
            btnSignup.classList.remove("disabled");
        }
        else {
            btnSignup.classList.add("disabled");
        }
    }
});
import "local://includes/routes/user.js";
import "local://includes/loading-utils.js";

var profileImg = document.getElementById("profile-img");
var btnUpdateProfile = document.getElementById("btn-update-profile");
var fileInp = document.getElementById("fileInput");
var form = document.getElementById("form-update-account");
var btnModify = document.getElementById("btn-modify");
var validationItemsElem = document.querySelectorAll(".pwd-validation__item");
var isFromSubscription = new URLSearchParams(window.location.search).get("session_id");
var planItemElem = document.querySelector(".plan__item");
var newPicture;
var accountUser = document.querySelector(".account__user");
var btnManagePlanElem = document.getElementById("btn-manage-plan");
var inpNewPassword = document.querySelector(".inp-new-password");

if (isFromSubscription) {
    ALERT_MOD.call({
        title: "Thank you !",
        withCheckmark: true,
        text: "You are now on the Basic plan paying monthly."
    });
}

if (USER) {
    form.elements["username"].value = USER.username;
    form.elements["email"].value = USER.email;

    if (USER.active_billing) {
        planItemElem.classList.add("show");
    }
    else {
        planItemElem.remove();
    }

    if (USER.picture_link) {
        profileImg.src = USER.picture_link;
    }
    else {
        profileImg.src = "/assets/images/no-profile.png";
    }

    if (accountUser) {
        accountUser.textContent = USER.username;
    }
}

btnUpdateProfile.addEventListener("click", function () {
    fileInp.click();
});

fileInp.addEventListener("change", function (e) {
    var files = this.files;

    if (files.length) {
        console.log(this.files[0].name, this.files[0]);

        btnModify.classList.remove("disabled");

        profileImg.src = window.URL.createObjectURL(this.files[0]);

        newPicture = this.files[0];
    }
}, false);

if (btnManagePlanElem) {
    btnManagePlanElem.addEventListener("click", function () {
        stripeCustomerPortal().then(cb => {
            if (cb && cb.success) {
                openInNewTab(cb.url);
            }
        })
    });    
}

function openInNewTab(href) {
    Object.assign(document.createElement('a'), {
        target: '_blank',
        href: href,
    }).click();
}

btnModify.addEventListener("click", function () {
    let oldText = this.textContent;
    this.classList.add("btn-clicked");
    this.disabled = true;

    if (newPicture) {
        var formData = new FormData();
        formData.append('file', newPicture);

        let promise = waitAtLeast(800, postUserPicture(formData));

        promise.then((cb) => {
            if (cb && cb.success) {
                this.classList.remove("btn-clicked");
                this.classList.add("btn-validated");
                this.textContent = "";

                setTimeout(() => {
                    this.classList.remove("btn-validated");
                    this.textContent = oldText;
                    this.disabled = false;
                }, 1500);

                getUser().then(cb => {
                    if (cb && cb.success) {
                        USER = cb.user;
                        setUserContent(USER);
                    }
                });
            }
            else {
                this.classList.remove("btn-clicked");
                this.classList.remove("btn-validated");
                this.textContent = oldText;
                this.disabled = false;

                ALERT_MOD.call({
                    title: "Information",
                    text: cb.code + " : " + cb.message
                });
            }
        });
    }

    if (form.elements["oldpassword"].value.length > 0 && form.elements["newpassword"].value.length > 0) {
        let promise = waitAtLeast(800, changePassword(form.elements["oldpassword"].value, form.elements["newpassword"].value));

        promise.then((cb) => {
            if (cb && cb.success) {
                this.classList.remove("btn-clicked");
                this.classList.add("btn-validated");
                this.textContent = "";

                setTimeout(() => {
                    this.classList.remove("btn-validated");
                    this.textContent = oldText;
                    this.disabled = false;
                }, 1500);
            }
            else {
                this.classList.remove("btn-clicked");
                this.classList.remove("btn-validated");
                this.textContent = oldText;
                this.disabled = false;

                ALERT_MOD.call({
                    title: "Information",
                    text: cb.code + " : " + cb.message
                });
            }
        });
    }

});

form.elements["newpassword"].addEventListener("keyup", function (e) {
    const regex1 = new RegExp(/.{8,}$/);                                               // at least 8 characters
    const regex2 = new RegExp(/(?=.*[0-9])/);                                          // at least one number
    const regex3 = new RegExp(/[ `!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~]/);             // at least one symbol

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

    checkSignupButton(isValidRegex1, isValidRegex2, isValidRegex3);
});

function checkSignupButton(isValidRegex1, isValidRegex2, isValidRegex3) {
    if (form.elements["oldpassword"].value.length > 0 && isValidRegex1 && isValidRegex2 && isValidRegex3) {
        btnModify.classList.remove("disabled");
        inpNewPassword.classList.add("is-success");
    }
    else {
        btnModify.classList.add("disabled");
        inpNewPassword.classList.remove("is-success");
    }
}

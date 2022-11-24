var domain = CONFIG.domain;

async function signup(email, username, password) {
    return fetch(domain + "/signup", {
        method: "POST",
        credentials: "include",
        body: JSON.stringify({
            "email": email,
            "username": username,
            "password": password
        })
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function signin(email, password) {
    return fetch(domain + "/cookiesignin", {
        method: "POST",
        credentials: "include",
        body: JSON.stringify({
            "email": email,
            "password": password
        })
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function signout() {
    return fetch(domain + "/user/signout", {
        method: "POST",
        credentials: "include",
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function forgetPassword(emailOrUsername) {
    return fetch(domain + "/forgotten-password", {
        method: "POST",
        body: JSON.stringify({
            "email": emailOrUsername,
        })
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function resetPassword(token, newPassword) {
    return fetch(domain + "/reset-password", {
        method: "POST",
        body: JSON.stringify({
            "token": token,
            "new_password": newPassword
        })
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function getUser() {
    return fetch(domain + "/user", {
        method: "GET",
        credentials: "include"
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function changePassword(oldPassword, newPassword) {
    return fetch(domain + "/user/change-password", {
        method: "POST",
        credentials: "include",
        body: JSON.stringify({
            "old_password": oldPassword,
            "new_password": newPassword
        })
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function getUsage(month, year) {
    return fetch(domain + "/user/usage?month=" + month + "&year=" + year, {
        method: "GET",
        credentials: "include"
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function postUserPicture(formData) {
    return fetch(domain + "/user/picture", {
        method: "POST",
        credentials: "include",
        body: formData
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

function getUserType(user) {
    switch (user.usermask) {
        // case 2:
        //     return "UserAdmin";
        // case 4:
        //     return "UserModerator";
        default:
            return user.userType;
    }
}


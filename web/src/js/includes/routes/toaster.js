
var domain = CONFIG.domain;

async function getToaster(id, donotrendermarkdown) {
    if (donotrendermarkdown) {
        return fetch(domain + "/toaster/" + id + "?donotrendermarkdown=true", {
            method: "GET",
            credentials: "include"
        })
            .then(response => response.json())
            .then(resp => Promise.resolve(resp))
            .catch(err => console.log(err));
    } else {
        return fetch(domain + "/toaster/" + id, {
            method: "GET",
            credentials: "include"
        })
            .then(response => response.json())
            .then(resp => Promise.resolve(resp))
            .catch(err => console.log(err));
    }
}

async function getToasterUsage(id) {
    return fetch(domain + "/toaster/usage/" + id, {
        method: "GET",
        credentials: "include"
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function getToasterRunningCount(id) {
    return fetch(domain + "/toaster/runningcount/" + id, {
        method: "GET",
        credentials: "include"
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function listToasterFiles(id) {
    return fetch(domain + "/toaster/listfiles/" + id, {
        method: "GET",
        credentials: "include"
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

// abspath is the file path and must begin with a leading slash
async function getToasterFile(id, abspath) {
    return fetch(domain + "/toaster/file/" + id + "/" + abspath, {
        method: "GET",
        credentials: "include"
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function getToasterExecutionLogs(id, exeID) {
    return fetch(domain + "/toaster/logs/" + id + "/" + exeID, {
        method: "GET",
        credentials: "include"
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function deleteToasters(...ids) {
    return fetch(domain + "/toaster", {
        method: "DELETE",
        credentials: "include",
        body: JSON.stringify({
            "toaster_ids": ids
        })
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function deleteToaster(id) {
    return fetch(domain + "/toaster/"+id, {
        method: "DELETE",
        credentials: "include"
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function listToasters() {
    return fetch(domain + "/toaster/list", {
        method: "GET",
        credentials: "include"
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

/*
    files is an array of File
    filepaths is an array of the string paths of the files

    Provide either {files and filepaths} or {gitURL, ...}
*/
async function createToaster(
    files,
    filepaths,
    gitURL,
    gitUsername,
    gitAccessToken,
    gitPassword,
    gitBranch,
    buildCmd,   // []string
    exeCmd,     // []string
    env,        // []string
    joinableForSeconds,
    maxConcurrentJoiners,
    timeoutSeconds,
    name,
    readme,
    keywords
) {
    var request = {
        "build_command": buildCmd,
        "execution_command": exeCmd,
        "environment_variables": env,
        "joinable_for_seconds": joinableForSeconds,
        "max_concurrent_joiners": maxConcurrentJoiners,
        "timeout_seconds": timeoutSeconds,
        "image": "ubuntu-20.04-nodejs-golang",
        "name": name,
        "readme": readme,
        "keywords": keywords
    };

    if (gitURL !== "") {
        request["git_url"] = gitURL;
        request["git_username"] = gitUsername;
        request["git_access_token"] = gitAccessToken;
        request["git_password"] = gitPassword;
        request["git_branch"] = gitBranch;

        return fetch(domain + "/toaster", {
            method: "POST",
            credentials: "include",
            body: JSON.stringify(request)
        })
            .then(response => response.json())
            .then(resp => Promise.resolve(resp))
            .catch(err => console.log(err));
    }
    else {
        var formData = new FormData();

        for (let i = 0; i < filepaths.length; i++) {
            formData.append("file", files[i], base32.encode(filepaths[i]));
        }

        formData.append("request", JSON.stringify(request));

        return fetch(domain + "/toaster", {
            method: "POST",
            credentials: "include",
            body: formData
        })
            .then(response => response.json())
            .then(resp => Promise.resolve(resp))
            .catch(err => console.log(err));
    }
}

async function updateToaster(
    id,
    files,
    filepaths,
    gitURL,
    gitUsername,
    gitAccessToken,
    gitPassword,
    gitBranch,
    buildCmd,       // []string
    exeCmd,         // []string
    env,            // []string
    joinableForSeconds,
    maxConcurrentJoiners,
    timeoutSeconds,
    name,
    readme,
    keywords,
) {
    var request = {
        "build_command": buildCmd,
        "execution_command": exeCmd,
        "environment_variables": env,
        "joinable_for_seconds": joinableForSeconds,
        "max_concurrent_joiners": maxConcurrentJoiners,
        "timeout_seconds": timeoutSeconds,
        "name": name,
        "readme": readme,
        "keywords": keywords
    };

    if (gitURL !== "") {
        request["git_url"] = gitURL;
        request["git_username"] = gitUsername;
        request["git_access_token"] = gitAccessToken;
        request["git_password"] = gitPassword;
        request["git_branch"] = gitBranch;

        return fetch(domain + "/toaster/" + id, {
            method: "PUT",
            credentials: "include",
            body: JSON.stringify(request)
        })
            .then(response => response.json())
            .then(resp => Promise.resolve(resp))
            .catch(err => console.log(err));
    }
    else {
        var formData = new FormData();
        for (let i = 0; i < filepaths.length; i++) {
            formData.append('file', files[i], base32.encode(filepaths[i]));
        }
        formData.append('request', JSON.stringify(request));

        return fetch(domain + "/toaster/" + id, {
            method: "PUT",
            credentials: "include",
            body: formData
        })
            .then(response => response.json())
            .then(resp => Promise.resolve(resp))
            .catch(err => console.log(err));
    }
}

async function postToasterPicture(formData, toasterId) {
    return fetch(domain + "/toaster/picture/" + toasterId, {
        method: "POST",
        credentials: "include",
        body: formData
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}


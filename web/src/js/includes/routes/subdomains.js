var domain = CONFIG.domain;

async function createSubDomain(name, optionalToasterID) {
    return fetch(domain + "/subdomain", {
        method: "POST",
        credentials: "include",
        body: JSON.stringify({
            "name": name,
            "toaster_id": optionalToasterID
        })
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function listSubDomains() {
    return fetch(domain + "/subdomain/list", {
        method: "GET",
        credentials: "include"
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function updateSubDomain(id, linkToasterID) {
    return fetch(domain + "/subdomain/" + id, {
        method: "PUT",
        credentials: "include",
        body: JSON.stringify({
            "toaster_id": linkToasterID
        })
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function getSubDomain(id) {
    return fetch(domain + "/subdomain/" + id, {
        method: "GET",
        credentials: "include"
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}


async function deleteSubDomain(id) {
    return fetch(domain + "/subdomain/" + id, {
        method: "DELETE",
        credentials: "include"
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

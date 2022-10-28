var domain = CONFIG.domain;

async function createCustomDomain(domains, linkedToasters) {
    return fetch(domain + "/customdomain", {
        method: "POST",
        credentials: "include",
        body: JSON.stringify({
            "subdomains": domains,
            "linked_toasters": linkedToasters
        })
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function listCustomDomains() {
    return fetch(domain + "/customdomain/list", {
        method: "GET",
        credentials: "include"
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function updateCustomDomain(id, domains, linkedToasters) {
    return fetch(domain + "/customdomain/" + id, {
        method: "PUT",
        credentials: "include",
        body: JSON.stringify({
            "subdomains": domains,
            "linked_toasters": linkedToasters
        })
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function verifyCustomDomain(id) {
    return fetch(domain + "/customdomain/verify/" + id, {
        method: "POST",
        credentials: "include"
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

async function getCustomDomain(id) {
    return fetch(domain + "/customdomain/" + id, {
        method: "GET",
        credentials: "include"
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}


async function deleteCustomDomain(id) {
    return fetch(domain + "/customdomain/" + id, {
        method: "DELETE",
        credentials: "include"
    })
        .then(response => response.json())
        .then(resp => Promise.resolve(resp))
        .catch(err => console.log(err));
}

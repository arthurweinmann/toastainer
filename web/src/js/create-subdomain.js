import "local://includes/routes/custom-domain.js";

var customDomainId = new URLSearchParams(window.location.search).get("from");
var isFromSubdomains = customDomainId ? true : false;
var h2RootDomain = document.querySelector(".h2-rootdomain");
var btnBack = document.querySelector(".back__cont");
var formCreateSubdomain = document.getElementById("form-create-subdomain");
var listDomains = [];

if (isFromSubdomains) {
    btnBack.style.display = "flex";
}
else {
    btnBack.style.display = "none";
}

if (customDomainId) {
    setCustomDomain();
}
else {
    window.location = "/custom-domains";
}

function setCustomDomain() {
    getCustomDomain(customDomainId).then(cb => {
        if (cb && cb.success) {
            let customDomain = cb.custom_domain;
            h2RootDomain.textContent = customDomain.root_domain;
            listDomains = customDomain.subdomains;

            if (!listDomains) {
                listDomains = [];
            }
            console.log("listDomains", listDomains);
        }
        else if (cb && !cb.success) {
            ALERT_MOD.call({
                title: "Information",
                text: "An error occurred while loading the domain.",
                buttons: [
                    {
                        text: "RETURN",
                        onClick: function () {
                            ALERT_MOD.close();
                            window.location = "/custom-domains";
                        }
                    },
                ]
            });
        }
        else {
            window.location = "/custom-domains";
        }
    });
}

btnBack.addEventListener("click", function () {
    history.back();
});

formCreateSubdomain.addEventListener("submit", function (e) {
    e.preventDefault();

    listDomains.push(this.elements["subdomainName"].value.trim().toLowerCase());

    updateCustomDomain(customDomainId, listDomains, undefined).then(cb => {
        if (cb && cb.success) {
            ALERT_MOD.call({
                title: "Success",
                text: "You have successfully created a subdomain.",
                buttons: [
                    {
                        text: "OK",
                        onClick: function () {
                            ALERT_MOD.close();
                            history.back();
                        }
                    },
                ]
            });
        }
        else if (cb && !cb.success) {
            ALERT_MOD.call({
                title: "Information",
                text: cb.code + " : " + cb.message
            });

            listDomains = listDomains.filter(dom => dom !== this.elements["subdomainName"].value.trim().toLowerCase());
        }
    });
});
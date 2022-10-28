import "local://includes/routes/custom-domain.js";
import "local://includes/loading-utils.js";

var customDomainId = new URLSearchParams(window.location.search).get("id");
var isFromCustomDomains = customDomainId ? true : false;
var h2RootDomain = document.querySelector(".h2-rootdomain");
var btnBack = document.querySelector(".back__cont");
var formEditDomain = document.getElementById("form-edit-domain");
var inpSubdomain = document.querySelector(".subdomain__inp");
var btnSubdomain = document.getElementById("btn-subdomain");
var tableBody = document.querySelector(".subdomain__table-body");
var subDomainsEmpty = document.querySelector(".subdomain__empty");
var subDomainsTable = document.querySelector(".subdmain__cont-table table");
var counterSubdomain = 0;
var listSubdomains = [];

if (isFromCustomDomains) {
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

btnBack.addEventListener("click", function () {
    window.location = "/custom-domains";
});

function setCustomDomain() {
    getCustomDomain(customDomainId).then(cb => {
        if (cb && cb.success) {
            let customDomain = cb.custom_domain;
            h2RootDomain.textContent = customDomain.root_domain;
            formEditDomain.elements["domainName"].value = customDomain.subdomains[0];

            for (let i = 1; i < customDomain.subdomains.length; i++) {
                addSubdomain(customDomain.subdomains[i]);
            }
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

function addSubdomain(value) {
    if (!listSubdomains.includes(value) && value.length > 0) {
        listSubdomains.push(value);
        inpSubdomain.value = "";
        counterSubdomain++;

        subDomainsEmpty.style.display = "none";
        subDomainsTable.style.display = "block";
        renderSubdomains(value);
    }
}

function renderSubdomains(value) {
    tableBody.insertAdjacentHTML("beforeend", `<tr id="subdomain-` + counterSubdomain + `">
        <td>` + value + `<span class="icon-x" for="subdomain-` + counterSubdomain + `"></span> </td>
    </tr>`);

    attachDeleteEvent(value);
}

function attachDeleteEvent(value) {
    document.querySelector("#subdomain-" + counterSubdomain + " .icon-x").addEventListener("click", function () {
        let forSubdomain = this.getAttribute("for");
        let subdomainToDelete = document.getElementById(forSubdomain);
        if (subdomainToDelete) {
            listSubdomains = listSubdomains.filter(subdomain => subdomain !== value);
            subdomainToDelete.remove();

            if (listSubdomains.length === 0) {
                subDomainsEmpty.style.display = "block";
                subDomainsTable.style.display = "none";
            }
        };
    });
}

inpSubdomain.addEventListener("keydown", function (e) {
    if (e.key === "Enter" || e.keyCode === 13) {
        let value = e.target.value.trim().toLowerCase();
        addSubdomain(value);
    }
});

btnSubdomain.addEventListener("click", function (e) {
    e.preventDefault();

    let value = inpSubdomain.value.trim().toLowerCase();
    addSubdomain(value);
});

/* avoid submit form on enter */
window.addEventListener("keydown", function (e) {
    if (e.key === "Enter") {
        e.preventDefault();
    }
});

formEditDomain.addEventListener("submit", function (e) {
    e.preventDefault();

    var domains = JSON.parse(JSON.stringify(listSubdomains));
    domains.splice(0, 0, this.elements["domainName"].value.trim().toLowerCase());
    
    ALERT_MOD.call({
        isLoading: true,
        loadingText: "Loading..."
    });

    let promise = waitAtLeast(800, updateCustomDomain(customDomainId, domains, undefined));

    promise.then((cb) => {
        if (cb && cb.success) {
            ALERT_MOD.call({
                title: "Success",
                withCheckmark: true, 
                text: "You have successfully updated a domain.",
                buttons: [
                    {
                        text: "OK",
                        onClick: function () {
                            ALERT_MOD.close();
                            window.location = "/custom-domains";
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
        }
    });
});
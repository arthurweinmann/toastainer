import "local://includes/routes/custom-domain.js";
import "local://includes/loading-utils.js";

var isFromCustomDomains = new URLSearchParams(window.location.search).get("from") ? true : false;
var btnBack = document.querySelector(".back__cont");
var formCreateDomain = document.getElementById("form-create-domain");
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


btnBack.addEventListener("click", function () {
    window.location = "/custom-domains";
});

function addSubdomain(value) {
    // TODO add regex check valid subdomain
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

formCreateDomain.addEventListener("submit", function (e) {
    e.preventDefault();

    var domains = JSON.parse(JSON.stringify(listSubdomains));
    domains.splice(0, 0, this.elements["domainName"].value.trim().toLowerCase());

    ALERT_MOD.call({
        isLoading: true,
        loadingText: "Loading..."
    });

    let promise = waitAtLeast(800, createCustomDomain(domains, undefined));

    promise.then((cb) => {
        if (cb && cb.success) {
            ALERT_MOD.call({
                title: "Success",
                text: "You have successfully created a domain.",
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
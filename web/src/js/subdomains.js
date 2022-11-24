import "local://includes/routes/subdomains.js";
import "local://includes/routes/toaster.js";
import "local://includes/modal-link-toaster.js";

var btnCreateDomain = document.getElementById("btn-create-subdomain");
var formCreateSubdomain = document.getElementById("form-create-subdomain");
var listDomainsElem = document.querySelector(".list-domains");
var gSubDomains = [];

setSubDomainsList();

function setSubDomainsList() {
    listSubDomains().then(cb => {
        if (cb && cb.success) {
            if (cb.subdomains) {
                gSubDomains = cb.subdomains;
                document.querySelector(".empty__zone").classList.remove("show");
                renderSubDomains();
            }
            else {
                gSubDomains = [];
                document.querySelector(".empty__zone").classList.add("show");
                renderSubDomains();
            }
        }
    });
}

function renderSubDomains() {
    listDomainsElem.innerHTML = ``;

    for (let i = 0; i < gSubDomains.length; i++) {
        listDomainsElem.insertAdjacentHTML("beforeend", getSubDomainItemHTML(gSubDomains[i]));
        attachEventDomainItem(gSubDomains[i].id);
    }
}

function getSubDomainItemHTML(subdomain) {
    let linked = "";

    if (subdomain.toaster_id && subdomain.toaster_id !== "") {
        linked = `
        <li class="domain__spec">
            <svg width="14" height="14" viewBox="0 0 9 9" fill="none"
                xmlns="http://www.w3.org/2000/svg">
                <g clip-path="url(#clip0_43_1052)">
                    <path
                        d="M0.781677 4.44019L1.434 3.78804C1.60697 3.61507 1.90474 3.73003 1.91371 3.97437C1.92506 4.28992 1.98244 4.60212 2.08404 4.90109C2.10142 4.95109 2.1044 5.00497 2.09266 5.05659C2.08091 5.10821 2.0549 5.1555 2.0176 5.19306L1.7875 5.42316C1.29531 5.91534 1.27949 6.71796 1.76728 7.21612C1.88445 7.3354 2.02408 7.4303 2.17811 7.49536C2.33213 7.56042 2.49751 7.59434 2.66471 7.59517C2.83191 7.59601 2.99762 7.56373 3.15229 7.50022C3.30696 7.4367 3.44752 7.34319 3.56588 7.22509L4.74642 6.04331C4.98359 5.80598 5.11682 5.48419 5.11682 5.14867C5.11682 4.81315 4.98359 4.49136 4.74642 4.25404C4.69042 4.19853 4.6296 4.14812 4.56467 4.10339C4.52842 4.07858 4.49849 4.04562 4.47728 4.00716C4.45606 3.9687 4.44415 3.9258 4.4425 3.88191C4.43873 3.78498 4.45512 3.68833 4.49064 3.59807C4.52617 3.50781 4.58004 3.42591 4.64886 3.35755L5.01906 2.98841C5.06613 2.94158 5.12821 2.91285 5.19437 2.90725C5.26053 2.90166 5.32655 2.91957 5.38082 2.95782C5.7003 3.1809 5.96715 3.47113 6.16266 3.80819C6.35817 4.14525 6.47762 4.52098 6.51263 4.90906C6.54765 5.29714 6.49739 5.68818 6.36537 6.05479C6.23335 6.42141 6.02277 6.75472 5.74838 7.03138L5.74205 7.03788L4.5608 8.21913C3.51894 9.26098 1.82388 9.26081 0.781501 8.21913C-0.260881 7.17745 -0.260178 5.48187 0.781677 4.44019Z"
                        fill="#9B9B9B" />
                    <path
                        d="M7.21303 3.57681C7.70521 3.08463 7.72103 2.28201 7.23324 1.78385C7.11605 1.66459 6.9764 1.56971 6.82236 1.50469C6.66832 1.43966 6.50294 1.40577 6.33574 1.40497C6.16854 1.40417 6.00283 1.43648 5.84818 1.50002C5.69352 1.56357 5.55298 1.65711 5.43465 1.77523L4.2541 2.95666C4.01693 3.19398 3.8837 3.51578 3.8837 3.8513C3.8837 4.18682 4.01693 4.50861 4.2541 4.74593C4.3101 4.80144 4.37093 4.85185 4.43586 4.89658C4.47207 4.92141 4.50197 4.95438 4.52315 4.99284C4.54433 5.0313 4.55622 5.07419 4.55785 5.11806C4.5617 5.21498 4.54536 5.31164 4.50986 5.40191C4.47437 5.49218 4.42049 5.57408 4.35166 5.64242L3.98147 6.01244C3.93437 6.05923 3.8723 6.08795 3.80615 6.09354C3.74 6.09914 3.67399 6.08125 3.61971 6.04303C3.30008 5.81996 3.0331 5.52969 2.8375 5.19255C2.64189 4.85541 2.52239 4.47957 2.48737 4.09138C2.45235 3.70318 2.50266 3.31203 2.63477 2.94533C2.76688 2.57863 2.9776 2.24526 3.25215 1.96859L3.25848 1.96209L4.43973 0.780837C5.48158 -0.261018 7.17664 -0.260842 8.21902 0.780837C9.2614 1.82252 9.26088 3.51775 8.21902 4.56013L7.5667 5.21228C7.39373 5.38525 7.09596 5.27029 7.08699 5.02595C7.07564 4.7104 7.01826 4.3982 6.91666 4.09924C6.89928 4.04923 6.89629 3.99535 6.90804 3.94373C6.91979 3.89211 6.9458 3.84482 6.9831 3.80726L7.21303 3.57681Z"
                        fill="white" />
                </g>
                <defs>
                    <clipPath id="clip0_43_1052">
                        <rect width="9" height="9" fill="white" />
                    </clipPath>
                </defs>
            </svg>
            <span class="domain__spec-label">Linked</span>
            <span class="displaynone linkedToasterID">` + subdomain.toaster_id + `</span>
        </li>
        `;
    }

    return `<li id="domain-` + subdomain.id + `" class="domain-item">
    <div class="domain__left">
        <div class="domain__name">
            <div class="ellipse"></div>
            <span class="spanDomainName">`+ subdomain.name + `</span>
        </div>
        <ul class="domain__list-specs">`+ linked + `</ul>
        </div>
        <div class="domain__cont-action">
            <div class="domain__action linkToaster">
                <span class="icon-link-2"></span>
                <span class="domain__action-label">Link</span>
            </div>
            <div class="domain__action delete">
                <span class="icon-trash"></span>
                <span class="domain__action-label">Delete</span>
            </div>
        </div>
    </li>`;
}

function attachEventDomainItem(subdomainid) {
    var domainElem = document.getElementById("domain-" + subdomainid);

    domainElem.querySelector(".domain__action.linkToaster").addEventListener("click", function (e) {
        e.preventDefault();
        e.stopPropagation();

        var linkedToasterIDContainer = document.getElementById("domain-" + subdomainid).querySelector(".linkedToasterID");
        var linkedToaster = "";
        if (linkedToasterIDContainer) {
            linkedToaster = linkedToasterIDContainer.textContent;
        }

        handleLinkToaster(subdomainid, linkedToaster);
    });

    domainElem.querySelector(".domain__action.delete").addEventListener("click", function (e) {
        e.preventDefault();
        e.stopPropagation();
        handleDeleteSubDomain(subdomainid);
    });
}

function handleLinkToaster(subdomainid, linkedToaster) {
    listToasters().then(cb => {
        if (cb && cb.success) {
            gToasters = cb.toasters;

            MODAL_LINK_TOASTER.call({
                title: "Link Toaster",
                text: "Select Toaster to link to your subdomain.",
                linkedToaster: linkedToaster,
                toasters: gToasters,
                buttons: [
                    {
                        text: "CANCEL",
                        onClick: function () {
                            MODAL_LINK_TOASTER.close();
                        }
                    },
                    {
                        text: "LINK",
                        onClick: function () {
                            let toasterToLink = "";

                            document.querySelectorAll(".toaster__item input").forEach(checkbox => {
                                if (checkbox.checked) {
                                    toasterToLink = checkbox.value;
                                }
                            });

                            updateSubDomain(subdomainid, toasterToLink).then(cb => {
                                if (cb && cb.success) {
                                    MODAL_LINK_TOASTER.close();
                                    ALERT_MOD.call({
                                        title: "Information",
                                        text: "Your Toaster has been successfully linked."
                                    });
                                    setSubDomainsList();
                                }
                                else if (cb && !cb.success) {
                                    MODAL_LINK_TOASTER.close();
                                    ALERT_MOD.call({
                                        title: "Information",
                                        text: cb.code + " : " + cb.message
                                    });
                                }
                            });
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
}

function handleDeleteSubDomain(subdomainid) {
    ALERT_MOD.call({
        title: "Confirmation",
        text: "Are you sure to delete this subdomain ?",
        buttons: [
            {
                text: "CANCEL",
                onClick: function () {
                    ALERT_MOD.close();
                }
            },
            {
                text: "DELETE",
                onClick: function () {
                    ALERT_MOD.close();
                    deleteSubDomain(subdomainid).then(cb => {
                        if (cb && cb.success) {
                            setSubDomainsList();
                        }
                        else if (cb && !cb.success) {
                            ALERT_MOD.call({
                                title: "Information",
                                text: cb.code + " : " + cb.message
                            });
                        }
                    })
                }
            },
        ]
    });
}

formCreateSubdomain.addEventListener("submit", function (e) {
    e.preventDefault();

    createSubDomain(this.elements["subdomainName"].value.trim().toLowerCase(), "").then(cb => {
        if (cb && cb.success) {
            ALERT_MOD.call({
                title: "Success",
                text: "You have successfully created a subdomain.",
                buttons: [
                    {
                        text: "OK",
                        onClick: function () {
                            ALERT_MOD.close();
                            setSubDomainsList();
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
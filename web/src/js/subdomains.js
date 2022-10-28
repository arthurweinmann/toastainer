import "local://includes/routes/custom-domain.js";
import "local://includes/routes/toaster.js";
import "local://includes/modal-link-toaster.js";

var customDomainId = new URLSearchParams(window.location.search).get("domain");
var isFromCustomDomains = customDomainId ? true : false;
var h1RootDomain = document.querySelector(".h1-rootdomain");
var btnBack = document.querySelector(".back__cont");
var btnHowTo = document.querySelector(".btn-howto");
var dnsContent = document.querySelector(".dns__content");
var btnCreateSubdomain = document.getElementById("btn-create-subdomain");
var listSubdomainsElem = document.querySelector(".list-subdomains");
var verificationTokenInp = document.querySelector(".token__value");
var btnCopyToken = document.getElementById("btn-copy-id");
var btnCopyTooltip = document.getElementById("btn-copy-tooltip");
var timeoutTooltip;

var gCustomDomain;
var gToasters = [];

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

btnCreateSubdomain.addEventListener("click", function () {
    window.location = "/create-subdomain?from=" + customDomainId;
});

btnHowTo.addEventListener("click", function () {
    if (btnHowTo.classList.contains("show")) {
        btnHowTo.classList.remove("show");
        dnsContent.classList.remove("show");
    }
    else {
        btnHowTo.classList.add("show");
        dnsContent.classList.add("show");
    }
});

/* ===== copy ==== */
btnCopyToken.addEventListener("click", function (e) {
    e.preventDefault();

    clearTimeout(timeoutTooltip);
    btnCopyTooltip.classList.add("show");

    timeoutTooltip = setTimeout(() => {
        btnCopyTooltip.classList.remove("show");
    }, 800);

    verificationTokenInp.select();
    verificationTokenInp.setSelectionRange(0, 99999); /* For mobile devices */

    navigator.clipboard.writeText(verificationTokenInp.value);
});

/* ===== render & set custom domain ==== */
function setCustomDomain() {
    getCustomDomain(customDomainId).then(cb => {
        if (cb && cb.success) {
            if (cb.custom_domain) {
                gCustomDomain = cb.custom_domain;

                h1RootDomain.textContent = gCustomDomain.root_domain;
                verificationTokenInp.value = gCustomDomain.verification_token;

                renderDNSRecords(cb);

                if (gCustomDomain.subdomains && gCustomDomain.subdomains.length > 0) {
                    renderSubdomains(gCustomDomain.subdomains);
                }
                else {
                    document.querySelector(".empty__zone").classList.add("show");
                    renderSubdomains([]);
                }
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

function renderSubdomains(subdomains) {
    listSubdomainsElem.innerHTML = ``;

    for (let i = 0; i < subdomains.length; i++) {
        let isFirst = i === 0;
        listSubdomainsElem.insertAdjacentHTML("beforeend", getSubdomainItemHTML(subdomains[i], isFirst));
        attachEventSubdomainItem(subdomains[i]);
    }
}

function getSubdomainItemHTML(subdomain, isFirst) {
    return `<li id="subdomain-` + subdomain + `"  class="subdomain-item">
        <div class="subdomain__left">
            <div class="subdomain__name">
                <div class="ellipse"></div>
                <span>`+ subdomain.substring(4) + `</span>
            </div>
        </div>
        <div class="subdomain__cont-action">
            <div class="subdomain__action unlinkToaster`+ (!gCustomDomain.linked_toasters || (gCustomDomain.linked_toasters && !gCustomDomain.linked_toasters[subdomain]) ? " disabled" : "") + `">
                <svg
                    xmlns="http://www.w3.org/2000/svg"
                    width="16"
                    height="16"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="rgba(255,255,255,.5)"
                    stroke-width="2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                >
                    <path d="M18.84 12.25L20.56 10.54H20.54C21.4606 9.58604 21.9651 8.30573 21.9426 6.98018C21.9201 5.65462 21.3725 4.39217 20.42 3.47C19.4869 2.57019 18.2412 2.06739 16.945 2.06739C15.6488 2.06739 14.4031 2.57019 13.47 3.47L11.75 5.18" />
                    <path d="M5.17 11.75L3.46 13.46C2.53937 14.414 2.03492 15.6943 2.05742 17.0198C2.07992 18.3454 2.62752 19.6078 3.58 20.53C4.51305 21.4298 5.75876 21.9326 7.055 21.9326C8.35124 21.9326 9.59695 21.4298 10.53 20.53L12.24 18.82" />
                    <path d="M8 2V5" />
                    <path d="M2 8H5" />
                    <path d="M16 19V22" />
                    <path d="M19 16H22" />
                    <path d="M19 19L21 21" />
                    <path d="M3 3L5 5" />
                </svg>
                `+ (isFirst ? `<span class="subdomain__action-label">Unlink</span>` : ``) + `
            </div>
            <div class="subdomain__action linkToaster">
                <span class="icon-link-2"></span>
                `+ (isFirst ? `<span class="subdomain__action-label">Link</span>` : ``) + `
            </div>
            <div class="subdomain__action delete">
                <span class="icon-trash"></span>
                `+ (isFirst ? ` <span class="subdomain__action-label">Delete</span>` : ``) + `
            </div>
        </div>
    </li>`;
}

function attachEventSubdomainItem(domainId) {
    var subdomainElem = document.getElementById("subdomain-" + domainId);

    subdomainElem.querySelector(".subdomain__action.delete").addEventListener("click", function (e) {
        e.preventDefault();
        e.stopPropagation();

        handleDeleteDomains(domainId);
    });

    subdomainElem.querySelector(".subdomain__action.linkToaster").addEventListener("click", function (e) {
        e.preventDefault();
        e.stopPropagation();

        handleLinkToaster(domainId);
    });

    subdomainElem.querySelector(".subdomain__action.unlinkToaster").addEventListener("click", function (e) {
        e.preventDefault();
        e.stopPropagation();

        handleUnlinkToaster(domainId);
    });
}

function handleDeleteDomains(domainId) {
    let domains = gCustomDomain.subdomains.filter(dom => dom !== domainId);

    ALERT_MOD.call({
        title: "Confirmation",
        text: "Are you sure to delete this domain ?",
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

                    updateCustomDomain(customDomainId, domains).then(cb => {
                        if (cb && cb.success) {
                            setCustomDomain();
                        }
                        else if (cb && !cb.success) {
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

function handleLinkToaster(domainId) {
    listToasters().then(cb => {
        if (cb && cb.success) {
            gToasters = cb.toasters;

            MODAL_LINK_TOASTER.call({
                title: "Link Toaster",
                text: "Select Toaster to link to your custom domain.",
                linkedToaster: gCustomDomain.linked_toasters ? (gCustomDomain.linked_toasters[domainId] ? gCustomDomain.linked_toasters[domainId] : "") : "",
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
                            let toastersToLink = gCustomDomain.linked_toasters ? gCustomDomain.linked_toasters : {};

                            document.querySelectorAll(".toaster__item input").forEach(checkbox => {
                                if (checkbox.checked) {
                                    let toaster = gToasters.find(toaster => toaster.id === checkbox.value);
                                    toastersToLink[domainId] = toaster.id;
                                }
                            });

                            updateCustomDomain(gCustomDomain.id, gCustomDomain.subdomains, toastersToLink).then(cb => {
                                if (cb && cb.success) {
                                    MODAL_LINK_TOASTER.close();
                                    ALERT_MOD.call({
                                        title: "Information",
                                        text: "Your Toaster has been successfully linked."
                                    });
                                    setCustomDomain();
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

function handleUnlinkToaster(domainId) {
    var toastersLinked = gCustomDomain.linked_toasters ? gCustomDomain.linked_toasters : {};

    delete toastersLinked[domainId];

    updateCustomDomain(gCustomDomain.id, gCustomDomain.subdomains, toastersLinked).then(cb => {
        if (cb && cb.success) {
            MODAL_LINK_TOASTER.close();
            ALERT_MOD.call({
                title: "Information",
                text: "Your Toaster has been successfully unlinked."
            });

            setCustomDomain();
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

function renderDNSRecords(customDomain) {
    var tableBodyRecord = document.querySelector(".subdomain__table-body");
    var mobileRecordsList = document.querySelector(".mobile-record__list");
    let recordsHTML = `<tr>
        <td>${customDomain.ownership_check_txt_record_name}</td>
        <td>300</td>
        <td>TXT</td>
        <td>${customDomain.ownership_check_txt_record_value}</td>
    </tr>`;
    let recordsMobileHTML = `<li class="mobile-record__item">
        <span><b>Name</b> ${customDomain.ownership_check_txt_record_name}</span>
        <span><b>TTL</b> 300</span>
        <span><b>Type</b> TXT</span>
        <span><b>Value</b> ${customDomain.ownership_check_txt_record_value}</span>
    </li>`;

    tableBodyRecord.innerHTML = ``;
    mobileRecordsList.innerHTML = ``;

    for (let cname in customDomain.cnames_record) {
        recordsHTML += `<tr>
            <td>${cname}</td>
            <td>300</td>
            <td>CNAME</td>
            <td>${customDomain.cnames_record[cname]}</td>
        </tr>`;

        recordsMobileHTML += `<li class="mobile-record__item">
            <span><b>Name</b> ${cname}</span>
            <span><b>TTL</b> 300</span>
            <span><b>Type</b> CNAME</span>
            <span><b>Value</b> ${customDomain.cnames_record[cname]}</span>
        </li>`;
    }

    tableBodyRecord.innerHTML = recordsHTML;
    mobileRecordsList.innerHTML = recordsMobileHTML;
}
import "local://includes/routes/custom-domain.js";

var btnCreateDomain = document.getElementById("btn-create-domain");
var listDomainsElem = document.querySelector(".list-domains");
var gCustomDomains = [];

btnCreateDomain.addEventListener("click", function () {
    window.location = "/create-domain?from=custom-domains";
});

setCustomDomainsList();

function setCustomDomainsList() {
    listCustomDomains().then(cb => {
        if (cb && cb.success) {
            if (cb.custom_domains) {
                gCustomDomains = cb.custom_domains;
                renderDomains();
            }
            else {
                gCustomDomains = [];
                document.querySelector(".empty__zone").classList.add("show");
                renderDomains();
            }
        }
    });
}

function renderDomains() {
    listDomainsElem.innerHTML = ``;

    for (let i = 0; i < gCustomDomains.length; i++) {
        let isFirst = i === 0;
        listDomainsElem.insertAdjacentHTML("beforeend", getDomainItemHTML(gCustomDomains[i], isFirst));
        attachEventDomainItem(gCustomDomains[i].id);
    }
}

function getDomainItemHTML(domain, isFirst) {
    return `<li id="domain-` + domain.id + `" class="domain-item">
    <div class="domain__left">
        <div class="domain__name">
            <div class="ellipse"></div>
            <span>`+ domain.root_domain + `</span>
        </div>
        <ul class="domain__list-specs">
            <li class="domain__spec`+ (!domain.ssl ? " disabled" : "") + `">
                <svg width="16" height="16" viewBox="0 0 14 14" fill="none"
                    xmlns="http://www.w3.org/2000/svg">
                    <path
                        d="M7 1.18672C3.68629 1.18672 0.999194 3.78984 0.999194 7C0.999194 10.2102 3.68629 12.8133 7 12.8133C10.3137 12.8133 13.0008 10.2102 13.0008 7C13.0008 3.78984 10.3137 1.18672 7 1.18672ZM4.25081 4.8207C4.25081 3.35234 5.48427 2.15742 7 2.15742C8.51573 2.15742 9.74919 3.35234 9.74919 4.8207V5.54805C9.74919 5.68477 9.63911 5.79141 9.49798 5.79141H8.99839C8.85726 5.79141 8.74718 5.68477 8.74718 5.54805V4.8207C8.74718 2.57578 5.24718 2.57578 5.24718 4.8207V5.54805C5.24718 5.68477 5.1371 5.79141 4.99597 5.79141H4.49637C4.35524 5.79141 4.24516 5.68477 4.24516 5.54805V4.8207H4.25081ZM10.9996 10.3906C10.9996 10.6559 10.7738 10.8746 10.5 10.8746H3.5C3.22621 10.8746 3.0004 10.6559 3.0004 10.3906V6.51602C3.0004 6.25078 3.22621 6.03203 3.5 6.03203H10.5C10.7738 6.03203 10.9996 6.25078 10.9996 6.51602V10.3906ZM3.9996 6.63633V10.2703C3.9996 10.3387 3.94597 10.3906 3.8754 10.3906H3.62419C3.55363 10.3906 3.5 10.3387 3.5 10.2703V6.63633C3.5 6.56797 3.55363 6.51602 3.62419 6.51602H3.8754C3.94597 6.51602 3.9996 6.56797 3.9996 6.63633ZM7.99919 7.96797C7.99919 8.32344 7.79597 8.63516 7.4996 8.79922V9.66328C7.4996 9.8 7.38952 9.90664 7.24839 9.90664H6.74879C6.60766 9.90664 6.49758 9.8 6.49758 9.66328V8.79922C6.20121 8.63242 5.99798 8.32344 5.99798 7.96797C5.99798 7.4293 6.44395 7 6.99718 7C7.5504 7 7.99919 7.43203 7.99919 7.96797ZM7 0.21875C3.13306 0.21875 0 3.25391 0 7C0 10.7461 3.13306 13.7812 7 13.7812C10.8669 13.7812 14 10.7461 14 7C14 3.25391 10.8669 0.21875 7 0.21875ZM7 13.2973C3.41532 13.2973 0.499597 10.4727 0.499597 7C0.499597 3.52734 3.41532 0.702734 7 0.702734C10.5847 0.702734 13.5004 3.52734 13.5004 7C13.5004 10.4727 10.5847 13.2973 7 13.2973Z"
                        fill="#FFF500" />
                </svg>
                <span class="domain__spec-label">SSL</span>
            </li>
            <li class="domain__spec`+ (!domain.enabled ? " disabled" : "") + `">
                <svg width="14" height="14" viewBox="0 0 11 11" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <g clip-path="url(#clip0_56_1266)">
                    <path d="M10.8389 3.10272L5.68799 8.28217L4.125 6.71575L9.28361 1.53801C9.33427 1.48692 9.3945 1.44632 9.46086 1.41852C9.52722 1.39071 9.59841 1.37626 9.67035 1.37598C9.7423 1.3757 9.8136 1.3896 9.88017 1.41689C9.94674 1.44417 10.0073 1.48431 10.0583 1.535L10.0613 1.53801L10.8389 2.32026C10.9421 2.42431 11 2.56493 11 2.71149C11 2.85804 10.9421 2.99866 10.8389 3.10272Z" fill="#42AE5A"/>
                    <path d="M5.68631 8.2841L4.51391 9.46295C4.41157 9.56608 4.27246 9.62433 4.12717 9.62489C3.98188 9.62545 3.84232 9.56829 3.73918 9.46596L3.73618 9.46295L0.161175 5.86732C0.057936 5.76333 0 5.62274 0 5.4762C0 5.32966 0.057936 5.18907 0.161175 5.08508L0.938909 4.30283C1.04108 4.19996 1.17988 4.1418 1.32487 4.14112C1.46986 4.14043 1.6092 4.19728 1.71235 4.29918L1.71578 4.30283L5.68631 8.2841Z" fill="white"/>
                    </g>
                    <defs>
                    <clipPath id="clip0_56_1266">
                    <rect width="11" height="11" fill="white"/>
                    </clipPath>
                    </defs>
                    </svg>
                    
                <span class="domain__spec-label">Verified</span>
            </li>
            <li class="domain__spec`+ (!domain.linked_toasters || domain.linked_toasters.length === 0 ? " disabled" : "") + `">
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
            </li>
        </ul>
        </div>
        <div class="domain__cont-action">
            <div class="domain__action verify">
                <span class="icon-check"></span>
                `+ (isFirst ? `<span class="domain__action-label">Verify</span>` : ``) + `
            </div>
            <div class="domain__action edit">
                <span class="icon-edit-3"></span>
                `+ (isFirst ? `<span class="domain__action-label">Edit</span>` : ``) + `
            </div>
            <div class="domain__action delete">
                <span class="icon-trash"></span>
                `+ (isFirst ? `<span class="domain__action-label">Delete</span>` : ``) + `
            </div>
        </div>
    </li>`;
}

function attachEventDomainItem(domainId) {
    var domainElem = document.getElementById("domain-" + domainId);

    domainElem.addEventListener("click", function () {
        window.location = "/subdomains?domain=" + domainId;
    });

    domainElem.querySelector(".domain__action.verify").addEventListener("click", function (e) {
        e.preventDefault();
        e.stopPropagation();
        handleVerifyCustomDomain(domainId);
    });

    domainElem.querySelector(".domain__action.edit").addEventListener("click", function (e) {
        e.preventDefault();
        e.stopPropagation();
        window.location = "/edit-domain?id=" + domainId;
    });

    domainElem.querySelector(".domain__action.delete").addEventListener("click", function (e) {
        e.preventDefault();
        e.stopPropagation();
        handleDeleteCustomDomain(domainId);
    });
}

function handleDeleteCustomDomain(domainId) {
    ALERT_MOD.call({
        title: "Confirmation",
        text: "Are you sure to delete this domain ? You will lose all created subdomains.",
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
                    deleteCustomDomain(domainId).then(cb => {
                        if (cb && cb.success) {
                            setCustomDomainsList();
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

function handleVerifyCustomDomain(domainId) {
    verifyCustomDomain(domainId).then(cb => {
        if (cb && cb.success) {
            ALERT_MOD.call({
                withCheckmark: true,
                title: "Success",
                text: "Your custom domain has been successfully verified."
            });
            setCustomDomainsList();
        }
        else if (cb && !cb.success) {
            ALERT_MOD.call({
                title: "Information",
                text: cb.code + " : " + cb.message
            });
        }
    });
}
import "local://includes/routes/toaster.js";
import "local://includes/routes/custom-domain.js";
import "local://includes/counter.js";
import "local://includes/markdown.js";
import "local://includes/alert-step.js";
import "local://includes/modal-link-custom-domain.js";
import "local://includes/utils/date.js";

var formToaster = document.getElementById("form-toaster");
var btnCopyToasterId = document.getElementById("btn-copy-id");
var tooltipCopyId = document.getElementById("tooltip-copy-id");
var btnCopyRun = document.getElementById("btn-copy-run");
var keywordsListElem = document.getElementById("keywords-list");
var envsListElem = document.getElementById("environments-list");
var buildCmdListElem = document.getElementById("build-cmd-list");
var exeCmdListElem = document.getElementById("execution-cmd-list");
var wrapperActionElem = document.querySelector(".wrapper-action");
var btnSelectAction = document.getElementById("btn-select-action");
var btnSelectRun = document.getElementById("btn-select-run");
var actionsListElem = document.querySelector(".actions__list");
var runContentElem = document.querySelector(".run__content");
var wrapperRunElem = document.querySelector(".wrapper-run")
var toasterImgElem = document.querySelector(".toaster__image");

var markdownContent = document.getElementById("markdown-content");
var timeoutTooltipId;
var timeoutTooltipRun;
var toasterIdFromURL = new URLSearchParams(window.location.search).get("id");
var gToaster = null;
var gCustomDomains = [];

if (toasterIdFromURL) {
    setToaster(toasterIdFromURL);

    // listToasterFiles(toasterIdFromURL).then(cb => {
    //     console.log(cb);
    //     if (cb && cb.success) {
    //         getToasterFile(toasterIdFromURL, cb.files[0]).then(cb => {
    //             console.log("cb", cb);
    //         })
    //     }
    // });
}
else {
    ALERT_MOD.call({
        title: "Information",
        text: "Toaster does not exist or has been deleted.",
        buttons: [
            {
                text: "OK",
                onClick: function () {
                    ALERT_MOD.close();
                    window.location = "/my-toasters";
                }
            },
        ]
    });
}

function setToaster(toasterId) {
    getToasterStats(toasterId).then(cb => {
        if (cb && cb.success) {
            let stats = cb.stats;

            document.querySelector(".shimmer__toaster-stats").style.display = "none";
            document.querySelector(".stats").classList.add("show");

            counterAnim("#stat1", 0, stats.durationms ? stats.durationms : 0, 300, "MS");
            counterAnim("#stat2", 0, stats.cpus ? stats.cpus : 0, 300, "S");
            counterAnim("#stat3", 0, stats.ramgbs ? stats.ramgbs : 0, 300, "GB/S");
            counterAnim("#stat4", 0, stats.ingress ? stats.ingress : 0, 300, "B");
            counterAnim("#stat5", 0, stats.egress ? stats.egress : 0, 300, "B");
        }
    });

    getToasterRunningCount(toasterId).then(cb => {
        let toasterCountRunning = 0;

        document.querySelector(".toaster-currently-running").classList.add("show");

        if (cb && cb.success) {
            toasterCountRunning = cb.running;
        }

        document.getElementById("toaster-running").textContent = toasterCountRunning;
    });

    getToaster(toasterId).then(cb => {
        if (cb && cb.success) {
            document.querySelectorAll(".shimmer").forEach(shimmer => shimmer.style.display = "none");
            toasterImgElem.classList.add("show");
            document.querySelector(".toaster__id-inp").classList.add("show");
            document.querySelector(".crypt__wrapper").classList.add("show");
            document.querySelector(".section-readme").classList.add("show");
            document.querySelectorAll(".cont-inp").forEach(contInp => contInp.classList.add("show"));
            document.getElementById("run-toaster").value = "https://" + cb.domain;

            gToaster = cb.toaster;
 
            document.querySelector(".toaster__created").textContent = getShortDate(new Date(gToaster.created * 1000));
            document.querySelector(".toaster__modified").textContent = timeSince(new Date(gToaster.last_modified));

            if (gToaster.picture_link) {
                let svg = toasterImgElem.querySelector("svg");

                toasterImgElem.style.backgroundImage = `url('${gToaster.picture_link}')`;
                if (svg) {
                    svg.remove();
                }
            }
            else {
                let img =  toasterImgElem.querySelector("img");
                if (img) {
                    toasterImgElem.querySelector("img").remove();
                }
            }

            setInputValues(gToaster);
        }
        else {
            ALERT_MOD.call({
                title: "Information",
                text: "Toaster does not exist or has been deleted.",
                buttons: [
                    {
                        text: "OK",
                        onClick: function () {
                            ALERT_MOD.close();
                            window.location = "/my-toasters";
                        }
                    },
                ]
            });
        }
    });
}

function setInputValues(toaster) {
    document.getElementById("name-toaster").textContent = toaster.name ? toaster.name : getShortDate(new Date(toaster.created * 1000));

    if (toaster.keywords) {
        for (let i = 0; i < toaster.keywords.length; i++) {
            let keyword = toaster.keywords[i];
            keywordsListElem.insertAdjacentHTML("beforeend", `<li class="keyword__item">` + keyword + `</li>`);
        }
    }
    else {
        document.querySelector(".section-keywords").style.display = "none";
    }

    if (toaster.environment_variables) {
        for (let i = 0; i < toaster.environment_variables.length; i++) {
            let env = toaster.environment_variables[i];
            envsListElem.insertAdjacentHTML("beforeend", `<li class="keyword__item">` + env + `</li>`);
        }
    }
    else {
        document.querySelector(".section-envs").style.display = "none";
    }

    if (!toaster.build_command && !toaster.execution_command) {
        document.querySelector(".section-configuration").style.display = "none";
    }
    else {
        if (toaster.build_command) {
            for (let i = 0; i < toaster.build_command.length; i++) {
                let buildCmd = toaster.build_command[i];
                buildCmdListElem.insertAdjacentHTML("beforeend", `<li class="command__item">` + buildCmd + `</li>`);
            }
        }
        else {
            document.querySelector(".section-configuration .config__cmd-left").style.display = "none";
        }
    
        if (toaster.execution_command) {
            for (let i = 0; i < toaster.execution_command.length; i++) {
                let exeCmd = toaster.execution_command[i];
                exeCmdListElem.insertAdjacentHTML("beforeend", `<li class="command__item">` + exeCmd + `</li>`);
            }
        }
        else {
            document.querySelector(".section-configuration .config__cmd-right").style.display = "none";
        }
    }



    formToaster.elements["toaster-id"].value = toaster.id;
    formToaster.elements["joinable-for-seconds"].value = toaster.joinable_for_seconds;
    formToaster.elements["max-concurrent-joiners"].value = toaster.max_concurrent_joiners;
    formToaster.elements["timeout-seconds"].value = toaster.timeout_seconds;

    // readme
    if (toaster.readme) {
        markdownContent.innerHTML = marked.parse(toaster.readme);
    }
    else {
        document.querySelector(".section-readme").style.display = "none";
    }
}

/* ===== copy ==== */
btnCopyToasterId.addEventListener("click", function (e) {
    e.preventDefault();

    var copyText = document.querySelector(".toaster__id-inp");

    clearTimeout(timeoutTooltipId);
    tooltipCopyId.classList.add("show");

    timeoutTooltipId = setTimeout(() => {
        tooltipCopyId.classList.remove("show");
    }, 800);

    copyText.select();
    copyText.setSelectionRange(0, 99999); /* For mobile devices */

    navigator.clipboard.writeText(copyText.value);
});

btnCopyRun.addEventListener("click", function (e) {
    e.preventDefault();

    var copyText = document.getElementById("run-toaster");

    clearTimeout(timeoutTooltipRun);

    btnCopyRun.textContent = "Copied";
    btnCopyRun.classList.remove("icon-copy");
    btnCopyRun.classList.add("active");

    timeoutTooltipRun = setTimeout(() => {
        btnCopyRun.textContent = "";
        btnCopyRun.classList.add("icon-copy");
    }, 2000);

    copyText.select();
    copyText.setSelectionRange(0, 99999); /* For mobile devices */

    navigator.clipboard.writeText(copyText.value);
});
/* ==== actions ==== */
window.addEventListener("click", function (e) {
    if (!wrapperActionElem.contains(e.target)) {
        actionsListElem.classList.remove("show");
    }

    if (!wrapperRunElem.contains(e.target)) {
        runContentElem.classList.remove("show");
    }
});

btnSelectRun.addEventListener("click", function () {
    if (runContentElem.classList.contains("show")) {
        runContentElem.classList.remove("show");
    }
    else {
        runContentElem.classList.add("show");
    }
});

btnSelectAction.addEventListener("click", function () {
    if (actionsListElem.classList.contains("show")) {
        actionsListElem.classList.remove("show");
    }
    else {
        actionsListElem.classList.add("show");
    }
});

for (let action of actionsListElem.children) {
    action.addEventListener("click", function () {
        let typeAction = this.getAttribute("data-action");

        switch (typeAction) {
            case "edit":
                handleEditToaster();
                break;
            case "delete":
                handleDeleteToaster();
                break;
            case "link":
                handleLinkToaster();
                break;
            default:
        }
    });
}

function handleEditToaster() {
    window.location = "/edit-toaster?id=" + gToaster.id;
}

function handleDeleteToaster() {
    ALERT_MOD.call({
        title: "Confirmation",
        text: "Are you sure to delete this Toaster ?",
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

                    deleteToaster(gToaster.id).then(cb => {
                        if (cb && cb.success) {
                            ALERT_MOD.call({
                                title: "Information",
                                text: "Your Toaster has been successfully removed."
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
            },
        ]
    });
}

function handleLinkToaster() {
    listCustomDomains().then(cb => {
        if (cb && cb.success) {
            gCustomDomains = cb.custom_domains;

            if (gCustomDomains) {
                MODAL_LINK_CUSTOM_DOM.call({
                    title: "Link to custom domain",
                    text: "Select custom domains to link to your Toaster.",
                    toasterId: gToaster.id,
                    customDomains: gCustomDomains,
                    buttons: [
                        {
                            text: "CANCEL",
                            onClick: function () {
                                MODAL_LINK_CUSTOM_DOM.close();
                            }
                        },
                        {
                            text: "LINK",
                            onClick: function () {
                                let linkCustomDomain = {};
    
                                for (let i = 0; i < gCustomDomains.length; i++) {
                                    if (gCustomDomains[i].linked_toasters) {
                                        for (let subdomain in gCustomDomains[i].linked_toasters) {
                                            if (gCustomDomains[i].linked_toasters[subdomain] === gToaster.id) {
                                                linkCustomDomain[gCustomDomains[i]["id"]] = {
                                                    domains: gCustomDomains[i].subdomains,
                                                    rootDomain: gCustomDomains[i].root_domain,
                                                    toasters: gCustomDomains[i].linked_toasters
                                                };
                                            }
                                        }
                                    }
                                }
    
                                document.querySelectorAll(".domain__item input").forEach(checkbox => {
                                    let cDomainId = checkbox.getAttribute("data-custom-domain");
                                    let cDomain = gCustomDomains.find(customDomain => customDomain.id === cDomainId);
    
                                    if (checkbox.checked) {
                                        if (!linkCustomDomain[cDomainId] && cDomain) {
                                            linkCustomDomain[cDomainId] = {
                                                rootDomain: cDomain.root_domain,
                                                domains: cDomain.subdomains,
                                                toasters: cDomain.linked_toasters ? cDomain.linked_toasters : {}
                                            };
                                        }
    
                                        linkCustomDomain[cDomainId] = {
                                            ...linkCustomDomain[cDomainId],
                                            toasters: {
                                                ...linkCustomDomain[cDomainId].toasters,
                                                [checkbox.value]: gToaster.id
                                            }
                                        };
                                    }
                                    else if (cDomain.linked_toasters && cDomain.linked_toasters[checkbox.value] === gToaster.id) {
                                        delete linkCustomDomain[cDomainId].toasters[checkbox.value];
                                    }
                                });
    
                                handleUpdateDomain(1, linkCustomDomain);
                            }
                        },
                    ]
                });
            }
            else {
                ALERT_MOD.call({
                    title: "Information",
                    text: "You have no custom domains yet."
                });
            }
        }
        else if (cb && !cb.success) {
            ALERT_MOD.call({
                title: "Information",
                text: cb.code + " : " + cb.message
            });
        }
    });
}

/* ==== markdown ==== */
const typed = () => {
    let text = ""; //localStorage.markupValue || markupArea.value;
    const newText = marked.parse(text);

    markdownContent.innerHTML = newText;
};

typed();

/* ==== handle update domain step */
async function handleUpdateDomain(step, linkCustomDomain) {
    let nbSteps = Object.keys(linkCustomDomain).length;     // equal to nb of custom domain updates
    let cDomainId = Object.keys(linkCustomDomain)[step - 1];

    console.log(gCustomDomains);
    if (step <= nbSteps) {
        updateCustomDomain(cDomainId, linkCustomDomain[cDomainId].subdomains, linkCustomDomain[cDomainId].toasters).then(cb => {
            if (cb && cb.success) {
                ALERT_STEP_MOD.call({
                    title: linkCustomDomain[cDomainId].rootDomain,
                    step: step,
                    nbSteps: nbSteps,
                    withCheckmark: true, 
                    text: "Your Toaster has been successfully linked.",
                    buttons: [
                        {
                            text: step === nbSteps ? "OK" : "NEXT",
                            onClick: function () {
                                step++;
                                handleUpdateDomain(step, linkCustomDomain);
                                if (step > nbSteps) {
                                    ALERT_STEP_MOD.close();
                                    MODAL_LINK_CUSTOM_DOM.close();
                                }
                            }
                        },
                    ]
                });
            }
            else if (cb && !cb.success) {
                ALERT_STEP_MOD.call({
                    title: linkCustomDomain[cDomainId].rootDomain,
                    step: step,
                    nbSteps: nbSteps,
                    text: cb.code + " : " + cb.message,
                    buttons: [
                        {
                            text: step === nbSteps ? "OK" : "NEXT",
                            onClick: function () {
                                step++;
                                handleUpdateDomain(step, linkCustomDomain);
                                if (step > nbSteps) {
                                    ALERT_STEP_MOD.close();
                                    MODAL_LINK_CUSTOM_DOM.close();
                                }
                            }
                        },
                    ]
                });
            }
        });
    }
}
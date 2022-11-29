import "local://includes/routes/toaster.js";
import "local://includes/tabs.js";
import "local://includes/markdown.js";
import "local://includes/loading-utils.js";

// window.addEventListener("beforeunload", function (e) {
//     var confirmationMessage = "It looks like you have been editing something. If you leave before saving, your changes will be lost.";

//     (e || window.event).returnValue = confirmationMessage;      // Gecko + IE
//     return confirmationMessage;                                 // Gecko + Webkit, Safari, Chrome etc.
// });

/* filesArray is normalized as
for directory
[{
    name,
    type,
    files: [{ file, path }
]}]

for file
[{
    type, 
    file
}]
*/

/* Code part */
var filesListElem = document.querySelector(".file__list");
var dropZone = document.getElementById("drop-zone");
var fileInp = document.getElementById("fileInput");
var folderInp = document.getElementById("folderInput");
var draggingZone = document.querySelector(".is-dragging__cont");
var fileContent = document.querySelector(".file__content");
var btnUploadFiles = document.querySelector(".upload-files");
var btnUploadFolder = document.querySelector(".upload-folder");
var btnNextStep = document.querySelector(".btn-next-step");

/* Information part */
var formToaster = document.getElementById("form-edit-toaster");
var btnDeleteImgToaster = document.querySelector(".cont-toaster__image .icon-x");
var toasterNoImg = document.querySelector(".toaster__no-image");
var btnUploadToasterImg = document.querySelector(".cont-toaster__image");
var imgFileInput = document.getElementById("imgFileInput");
var keywordsListElem = document.getElementById("keywords-list");
var indexKeywords = 0;
var timeoutFlashKeyword;
var markdownContent = document.getElementById("markdown-content");
var markupArea = document.getElementById("markup");

var timeoutFlashEnv;
var timeoutFlashBuildCmd;
var timeoutFlashExeCmd;

/* Toaster */
var gToaster = null;
var imgToaster = null;
var filesArray = [];
var keywords = [];
var environments = [];
var buildCommands = [];
var executionCommands = [];

var envIndex = 0;
var buildCmdIndex = 0;
var exeCmdIndex = 0;

/* handle import type */

document.querySelectorAll(".import__item").forEach(btnImportType => {
    btnImportType.addEventListener("click", function (e) {
        e.preventDefault();

        let importToDisplayId = this.getAttribute("data-for");

        document.querySelector(".import__item.active").classList.remove("active");
        btnImportType.classList.add("active");

        document.querySelector(".import__cont-body.active").classList.remove("active");
        document.getElementById(importToDisplayId).classList.add("active");
    });
});

/* ======== Code part ======== */

var toasterIdFromURL = new URLSearchParams(window.location.search).get("id");

if (toasterIdFromURL) {
    setToaster(toasterIdFromURL);
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

btnNextStep.addEventListener("click", function (e) {
    e.preventDefault();
    showTab("information");
});

function setToaster(toasterId) {
    getToaster(toasterId, true).then(cb => {
        if (cb && cb.success) {
            gToaster = cb.toaster;
            setInputValues(gToaster);

            if (gToaster.picture_link) {
                btnUploadToasterImg.style.backgroundImage = `url('${gToaster.picture_link}')`;
                toasterNoImg.style.display = "none";
            }
            else {
            }
        }
        else {
            ALERT_MOD.call({
                title: "Information",
                text: cb.code + " : " + cb.message
            });
        }
    })
}

function setInputValues(toaster) {
    document.getElementById("name-toaster").textContent = toaster.name;

    if (toaster.keywords) { keywords = toaster.keywords; }
    if (toaster.environment_variables) { environments = toaster.environment_variables; }
    if (toaster.build_command) { buildCommands = toaster.build_command; }
    if (toaster.execution_command) { executionCommands = toaster.execution_command; }

    for (let i = 0; i < keywords.length; i++) {
        let keyword = keywords[i];
        addKeyword(keywordsListElem, "keyword", keyword);
    }

    /* set envs */
    var listEnvsElem = document.querySelector(".list-envs");
    envIndex = environments.length;

    for (let i = 0; i < environments.length; i++) {
        let env = environments[i].split("=");
        let envKey = env[0];
        let envValue = env[1];

        listEnvsElem.insertAdjacentHTML("beforeend", `<li id="env-${i}" class="cont-inp-line" data-index="${i}">
            <div class="cont-inp env-key">
                <label><span>Key<span class="label-required">*</span></span>
                    <input name="env-key-${i}" value="${envKey}"/>
                </label>
            </div>
            <div class="cont-inp env-value">
                <label><span>Value<span class="label-required">*</span></span>
                    <input name="env-value-${i}" value="${envValue}"/>
                </label>
            </div>
            <div class="cont-env-close close-${i}">
                <span class="icon-x"></span>
            </div>
        </li>`);

        document.querySelector(`.cont-env-close.close-${i}`).addEventListener("click", function () {
            this.parentElement.remove();
        });
    }



    /* set build commands */
    var listBuildsElem = document.querySelector(".list-build-cmds");
    buildCmdIndex = buildCommands.length;

    for (let i = 0; i < buildCommands.length; i++) {
        listBuildsElem.insertAdjacentHTML("beforeend", `<li id="build-${i}" class="cont-inp-line" data-index="${i}">
            <div class="cont-inp build-value">
                <label><span>Value<span class="label-required">*</span></span>
                    <input name="build-value-${i}" value="${buildCommands[i]}"/>
                </label>
            </div>
            <div class="cont-build-close close-${i}">
                <span class="icon-x"></span>
            </div>
        </li>`);

        document.querySelector(`.cont-build-close.close-${i}`).addEventListener("click", function () {
            this.parentElement.remove();
        });
    }

    /* set exe commands */
    var listExecsElem = document.querySelector(".list-exe-cmds");
    exeCmdIndex = executionCommands.length;

    for (let i = 0; i < executionCommands.length; i++) {
        listExecsElem.insertAdjacentHTML("beforeend", `<li id="exe-${i}" class="cont-inp-line" data-index="${i}">
                   <div class="cont-inp exe-value">
                       <label><span>Value<span class="label-required">*</span></span>
                           <input name="exe-value-${i}" value="${executionCommands[i]}"/>
                       </label>
                   </div>
                   <div class="cont-exe-close close-${i}">
                       <span class="icon-x"></span>
                   </div>
               </li>`);

        document.querySelector(`.cont-exe-close.close-${i}`).addEventListener("click", function () {
            this.parentElement.remove();
        });
    }

    formToaster.elements["name"].value = toaster.name;
    formToaster.elements["toaster-id"].value = toaster.id;
    formToaster.elements["joinable-for-seconds"].value = toaster.joinable_for_seconds ? toaster.joinable_for_seconds : 0;
    formToaster.elements["max-concurrent-joiners"].value = toaster.max_concurrent_joiners ? toaster.max_concurrent_joiners : 0;
    formToaster.elements["timeout-seconds"].value = toaster.timeout_seconds ? toaster.timeout_seconds : 0;
    formToaster.elements["read-me"].value = toaster.readme ? toaster.readme : "";

    if (toaster.readme) {
        typed();
    }
}

btnUploadFiles.addEventListener("click", function () {
    fileInp.click();
});

btnUploadFolder.addEventListener("click", function () {
    folderInp.click();
});

fileInp.addEventListener("change", function (e) {
    var files = this.files;

    if (files.length) {
        for (let i = 0; i < files.length; i++) {
            let isAlreadyExist = filesArray.findIndex(f => (f.file && f.file.name) === files[i].name) > -1;
            if (!isAlreadyExist) {
                filesArray.push({
                    type: "file",
                    file: files[i]
                });
            }
        }
    }

    renderFilesListElem();
}, false);

folderInp.addEventListener("change", function (e) {
    var folders = this.files;

    if (folders.length) {
        let filesDir = [];
        let folderName = folders[0].webkitRelativePath.split("/")[0];
        let isAlreadyExist = filesArray.findIndex(f => f.name === folderName) > -1;

        for (let i = 0; i < folders.length; i++) {
            let objFile = {
                path: folders[i].webkitRelativePath,
                file: folders[i]
            }
            filesDir.push(objFile);
        }

        if (!isAlreadyExist) {
            filesArray.push({
                name: folderName,
                type: "folder",
                files: filesDir
            });
        }
    }

    renderFilesListElem();
}, false);

document.getElementById("drop-zone").addEventListener("drop", async function (ev) {
    filesListElem.innerHTML = ``;

    // Prevent default behavior (Prevent file from being opened)
    ev.preventDefault();

    if (ev.dataTransfer.items) {
        // Use DataTransferItemList interface to access the file(s)
        for (let i = 0; i < ev.dataTransfer.items.length; i++) {
            // If dropped items aren't files, reject them
            if (ev.dataTransfer.items[i].kind === "file") {
                let file = ev.dataTransfer.items[i].getAsFile();

                if (!file.type && file.size % 4096 == 0) {      // folder
                    var fileEntries = await getAllFileEntries(ev.dataTransfer.items);
                    let filesDir = [];

                    for (let i = 0; i < fileEntries.length; i++) {
                        let fileConverted = await getFile(fileEntries[i]);
                        let objFile = {
                            path: fileEntries[i].fullPath.substring(1),
                            file: fileConverted
                        }
                        filesDir.push(objFile);
                    }

                    let isAlreadyExist = filesArray.findIndex(f => f.name === file.name) > -1;

                    if (!isAlreadyExist) {
                        filesArray.push({
                            name: file.name,
                            type: "folder",
                            files: filesDir
                        });
                    }
                }
                else {          // file
                    let isAlreadyExist = filesArray.findIndex(f => (f.file && f.file.name) === file.name) > -1;

                    if (!isAlreadyExist) {
                        filesArray.push({
                            type: "file",
                            file: file
                        });
                    }
                }
            }
        }
    }
    else {
        // Use DataTransfer interface to access the file(s)
        for (var i = 0; i < ev.dataTransfer.files.length; i++) {
            let file = ev.dataTransfer.files[i];
            let isAlreadyExist = filesArray.findIndex(f => f.file.name === file.name) > -1;

            if (!isAlreadyExist) {
                filesArray.push({
                    type: "file",
                    file: file
                });
            }
        }
    }

    dropZone.classList.remove("is-dragging");
    draggingZone.classList.remove("show");
    fileContent.classList.add("show");

    if (filesArray.length > 0) {
        filesListElem.classList.add("show");
    }
    else {
        filesListElem.classList.remove("show");
    }

    renderFilesListElem();
});

function renderFilesListElem() {
    filesListElem.innerHTML = ``;
    dropZone.style.height = "initial";

    for (let i = 0; i < filesArray.length; i++) {
        if (filesArray[i].type === "folder") {
            filesListElem.innerHTML += getFolderHTML(filesArray[i]);
        }
        else {
            filesListElem.innerHTML += getFileHTML(filesArray[i]);
        }
    }

    if (filesArray.length > 0) {
        filesListElem.classList.add("show");
    }
    else {
        filesListElem.classList.remove("show");
    }

    // adapt height dropzone
    dropZone.style.height = dropZone.offsetHeight + "px";

    // attach delete file events
    document.querySelectorAll(".file__item .icon-x").forEach(btnDeleteFile => {
        btnDeleteFile.addEventListener("click", function (e) {
            e.preventDefault();
            e.stopPropagation();

            let fileName = this.getAttribute("file-value");
            let fileType = this.getAttribute("file-type");
            let dt = new DataTransfer();

            if (fileType === "file") {
                for (let i = 0; i < fileInp.files.length; i++) {
                    let file = fileInp.files[i]
                    if (fileName !== file.name) dt.items.add(file) // here you exclude the file. thus removing it.
                }

                fileInp.files = dt.files;
            }
            else {
                folderInp.files = null;
            }

            filesArray = filesArray.filter(objFile => {
                if (objFile.type === "folder") return objFile.name !== fileName
                return objFile.file.name !== fileName
            });
            renderFilesListElem();
        });
    });
}

function getFileHTML(fileObj) {
    return `<li class="file__item">
        <span class="file__label">`+ fileObj.file.name + `</span>
        <span class="file__infos">` + formatSizeUnits(fileObj.file.size) + ` <span>&#183;</span> ` + fileObj.file.name.split(".").pop() + `</span>
        <span class="icon-x" file-type="file" file-value="`+ fileObj.file.name + `"></span>
    </li>`;
}

function getFolderHTML(folderObj) {
    return `<li class="file__item">
        <span class="file__label"><span class="icon-folder"></span>`+ folderObj.name + ` <span>&#183;</span> <span>` + folderObj.files.length + ` éléments</span></span>
        <span class="icon-x" file-type="folder" file-value="`+ folderObj.name + `"></span>
    </li>`;
}

// avoid opening file in new tab browser
document.getElementById("drop-zone").addEventListener("dragover", function (e) {
    e.preventDefault();
});

// show drag&drop zone 
document.getElementById("drop-zone").addEventListener("dragenter", function (e) {
    e.preventDefault();
    if (e.currentTarget.contains(e.relatedTarget)) return;
    dropZone.classList.add("is-dragging");
    draggingZone.classList.add("show");
    fileContent.classList.remove("show");
    filesListElem.classList.remove("show");
});

// hide drag&drop zone 
document.getElementById("drop-zone").addEventListener("dragleave", function (e) {
    e.preventDefault();
    if (e.currentTarget.contains(e.relatedTarget)) return;
    dropZone.classList.remove("is-dragging");
    draggingZone.classList.remove("show");
    fileContent.classList.add("show");

    if (filesArray.length > 0) {
        filesListElem.classList.add("show");
    }
    else {
        filesListElem.classList.remove("show");
    }
});

function formatSizeUnits(bytes) {
    if (bytes >= 1073741824) { bytes = (bytes / 1073741824).toFixed(2) + " GB"; }
    else if (bytes >= 1048576) { bytes = (bytes / 1048576).toFixed(2) + " MB"; }
    else if (bytes >= 1024) { bytes = (bytes / 1024).toFixed(2) + " KB"; }
    else if (bytes > 1) { bytes = bytes + " bytes"; }
    else if (bytes == 1) { bytes = bytes + " byte"; }
    else { bytes = "0 bytes"; }
    return bytes;
}

/* handle directory and contents drag & drop */
// Drop handler function to get all files
async function getAllFileEntries(dataTransferItemList) {
    let fileEntries = [];
    // Use BFS to traverse entire directory/file structure
    let queue = [];
    // Unfortunately dataTransferItemList is not iterable i.e. no forEach
    for (let i = 0; i < dataTransferItemList.length; i++) {
        queue.push(dataTransferItemList[i].webkitGetAsEntry());
    }

    while (queue.length > 0) {
        let entry = queue.shift();
        if (entry.isFile) {
            fileEntries.push(entry);
        }
        else if (entry.isDirectory) {
            queue.push(...await readAllDirectoryEntries(entry.createReader()));
        }
    }
    return fileEntries;
}

// Get all the entries (files or sub-directories) in a directory 
// by calling readEntries until it returns empty array
async function readAllDirectoryEntries(directoryReader) {
    let entries = [];
    let readEntries = await readEntriesPromise(directoryReader);
    while (readEntries.length > 0) {
        entries.push(...readEntries);
        readEntries = await readEntriesPromise(directoryReader);
    }
    return entries;
}

// Wrap readEntries in a promise to make working with readEntries easier
// readEntries will return only some of the entries in a directory
// e.g. Chrome returns at most 100 entries at a time
async function readEntriesPromise(directoryReader) {
    try {
        return await new Promise((resolve, reject) => {
            directoryReader.readEntries(resolve, reject);
        });
    }
    catch (err) {
        console.log(err);
    }
}

async function getFile(fileEntry) {
    try {
        return await new Promise((resolve, reject) => fileEntry.file(resolve, reject));
    }
    catch (err) {
        console.log(err);
    }
}


/* ======== Information part ======== */

/* toaster img */
btnUploadToasterImg.addEventListener("click", function () {
    imgFileInput.click();
});

imgFileInput.addEventListener("change", function (e) {
    var files = this.files;

    if (files.length) {
        btnUploadToasterImg.style.backgroundImage = `url('${window.URL.createObjectURL(this.files[0])}')`;
        btnDeleteImgToaster.style.display = "block";
        toasterNoImg.style.display = "none";
        imgToaster = this.files[0];

    }
}, false);


btnDeleteImgToaster.addEventListener("click", function (e) {
    e.stopPropagation();

    btnUploadToasterImg.style.backgroundImage = "";
    imgFileInput.value = null;
    btnDeleteImgToaster.style.display = "none";
    toasterNoImg.style.display = "flex";
    imgToaster = null;
});

/* handle keywords */
function addKeyword(listElem, name, value) {
    if (name === "keyword") {
        listElem.insertAdjacentHTML("beforeend", `<li id="` + name + `-` + indexKeywords + `" class="keyword__item">` + value + ` <span class="icon-x" key-value="` + value + `"></span></li>`);
        attachDeleteKeyword(name, document.getElementById(name + "-" + indexKeywords));
        indexKeywords++;
    }
}

function removeKeyword(itemElem) {
    itemElem.remove();
}

function attachDeleteKeyword(name, itemElem) {
    var btnDelete = itemElem.querySelector(".icon-x");

    btnDelete.addEventListener("click", function () {
        let value = this.getAttribute("key-value");

        if (name === "keyword") {
            keywords = keywords.filter(k => k !== value);
        }
        else if (name === "environment") {
            environments = environments.filter(k => k !== value);
        }

        removeKeyword(itemElem);
    });
}

keywordsListElem.querySelector(".inp-add-keywords input").addEventListener("keydown", function (e) {
    if (e.key === "Enter" || e.keyCode === 13) {
        e.preventDefault();

        let name = this.name;
        let value = e.target.value.trim().toLowerCase();

        if (!keywords.includes(value) && value.length > 0) {
            keywords.push(value);
            addKeyword(keywordsListElem, name, value);
        }
        else {
            highlightExistingKeyword(keywordsListElem, value);
        }

        this.value = "";
    }
});

function highlightExistingKeyword(listElem, value) {
    clearTimeout(timeoutFlashKeyword);

    listElem.querySelectorAll(".keyword__item:not(.inp-add-keywords) .icon-x").forEach(item => {
        let itemVal = item.getAttribute("key-value");
        item.parentElement.classList.remove("flash");

        if (itemVal === value.trim().toLowerCase()) {
            item.parentElement.classList.add("flash");

            timeoutFlashKeyword = setTimeout(() => {
                item.parentElement.classList.remove("flash");
            }, 1000);
        }
    })
}

document.getElementById("btn-add-env").addEventListener("click", function (e) {
    e.preventDefault();

    var listEnvsElem = document.querySelector(".list-envs");
    var lastEnvLine = listEnvsElem.querySelector("li:last-child");

    clearTimeout(timeoutFlashEnv);

    if (!lastEnvLine || (lastEnvLine.querySelector(".env-key input").value.length > 0 && lastEnvLine.querySelector(".env-value input").value.length > 0)) {
        listEnvsElem.insertAdjacentHTML("beforeend", `<li id="env-${envIndex}" class="cont-inp-line" data-index="${envIndex}">
            <div class="cont-inp env-key">
                <label><span>Key<span class="label-required">*</span></span>
                    <input name="env-key-${envIndex}" />
                </label>
            </div>
            <div class="cont-inp env-value">
                <label><span>Value<span class="label-required">*</span></span>
                    <input name="env-value-${envIndex}" />
                </label>
            </div>
            <div class="cont-env-close close-${envIndex}">
                <span class="icon-x"></span>
            </div>
        </li>`);

        document.querySelector(`.cont-env-close.close-${envIndex}`).addEventListener("click", function () {
            this.parentElement.remove();
        });
    }
    else {

        lastEnvLine.classList.add("required");

        timeoutFlashEnv = setTimeout(function () {
            lastEnvLine.classList.remove("required");
        }, 1000);
    }

    envIndex++;
});

document.getElementById("btn-add-build-cmd").addEventListener("click", function (e) {
    e.preventDefault();

    var listBuildsElem = document.querySelector(".list-build-cmds");
    var lastBuildLine = listBuildsElem.querySelector("li:last-child");

    clearTimeout(timeoutFlashBuildCmd);

    if (!lastBuildLine || lastBuildLine.querySelector(".build-value input").value.length > 0) {
        listBuildsElem.insertAdjacentHTML("beforeend", `<li id="build-${buildCmdIndex}" class="cont-inp-line" data-index="${buildCmdIndex}">
            <div class="cont-inp build-value">
                <label><span>Value<span class="label-required">*</span></span>
                    <input name="build-value-${buildCmdIndex}" />
                </label>
            </div>
            <div class="cont-build-close close-${buildCmdIndex}">
                <span class="icon-x"></span>
            </div>
        </li>`);

        document.querySelector(`.cont-build-close.close-${buildCmdIndex}`).addEventListener("click", function () {
            this.parentElement.remove();
        });
    }
    else {
        lastBuildLine.classList.add("required");

        timeoutFlashBuildCmd = setTimeout(function () {
            lastBuildLine.classList.remove("required");
        }, 1000);
    }

    buildCmdIndex++;
});

document.getElementById("btn-add-exe-cmd").addEventListener("click", function (e) {
    e.preventDefault();

    var listExecsElem = document.querySelector(".list-exe-cmds");
    var lastExeLine = listExecsElem.querySelector("li:last-child");

    clearTimeout(timeoutFlashExeCmd);

    if (!lastExeLine || lastExeLine.querySelector(".exe-value input").value.length > 0) {
        listExecsElem.insertAdjacentHTML("beforeend", `<li id="exe-${exeCmdIndex}" class="cont-inp-line" data-index="${exeCmdIndex}">
            <div class="cont-inp exe-value">
                <label><span>Value<span class="label-required">*</span></span>
                    <input name="exe-value-${exeCmdIndex}" />
                </label>
            </div>
            <div class="cont-exe-close close-${exeCmdIndex}">
                <span class="icon-x"></span>
            </div>
        </li>`);

        document.querySelector(`.cont-exe-close.close-${exeCmdIndex}`).addEventListener("click", function () {
            this.parentElement.remove();
        });
    }
    else {
        lastExeLine.classList.add("required");

        timeoutFlashExeCmd = setTimeout(function () {
            lastExeLine.classList.remove("required");
        }, 1000);
    }

    exeCmdIndex++;
});


/* handle readme & markdown */
markupArea.value = localStorage.markupValue ? localStorage.markupValue : "";

const typed = () => {
    let text = localStorage.markupValue || markupArea.value;
    const newText = marked.parse(text);

    markdownContent.innerHTML = newText;
    return markupArea.value;
};

typed();

markup.addEventListener("keyup", () => {
    localStorage.setItem("markupValue", typed());
    typed();
});

document.querySelectorAll(".cont-readme h3").forEach(tab => {
    tab.addEventListener("click", function () {
        let activeTab = document.querySelector(".cont-readme h3.active");
        let tabToHide = document.getElementById(activeTab.getAttribute("data-tab"));
        let tabToShow = document.getElementById(tab.getAttribute("data-tab"));

        activeTab.classList.remove("active");
        tab.classList.add("active");
        tabToHide.style.display = "none";
        tabToShow.style.display = "block";
    });
});

/* handle edit toaster */
/* avoid submit form on enter */
window.addEventListener("keydown", function (e) {
    if (e.key === "Enter" && !markupArea.contains(e.target)) {
        e.preventDefault();
    }
});

formToaster.addEventListener("submit", function (e) {
    e.preventDefault();

    var elements = this.elements;
    var filesAndFilePaths = toFilesAndFilesPath(filesArray);
    let joinableForSeconds = elements["joinable-for-seconds"].value.trim();
    let maxConcurrentJoiners = elements["max-concurrent-joiners"].value.trim();
    let timeoutSeconds = elements["timeout-seconds"].value.trim();
    var formDataImage = new FormData();
    formDataImage.append('file', imgToaster);

    // handle environments
    environments = [];
    document.querySelectorAll(".list-envs .cont-inp-line").forEach(envLineElem => {
        let envKey = envLineElem.querySelector(".env-key input").value.trim();
        let envValue = envLineElem.querySelector(".env-value input").value.trim();

        if (envKey.length > 0 && envValue.length > 0) {
            environments.push(`${envKey}=${envValue}`);
        }
    });

    // handle build commands
    buildCommands = [];
    document.querySelectorAll(".list-build-cmds .cont-inp-line").forEach(buildLineElem => {
        let buildValue = buildLineElem.querySelector(".build-value input").value.trim();

        if (buildValue.length > 0) {
            buildCommands.push(buildValue);
        }
    });

    // handle execution commands
    executionCommands = [];
    document.querySelectorAll(".list-exe-cmds .cont-inp-line").forEach(exeLineElem => {
        let exeValue = exeLineElem.querySelector(".exe-value input").value.trim();

        if (exeValue.length > 0) {
            executionCommands.push(exeValue);
        }
    });

    var toaster = {
        id: gToaster.id,

        // code
        files: filesAndFilePaths.files,
        filepaths: filesAndFilePaths.filePaths,
        gitUsername: elements["gitusername"].value.trim(),
        gitPassword: elements["gitpassword"].value.trim(),
        gitURL: elements["url"].value.trim(),
        gitBranch: elements["branch-ref"].value.trim(),
        gitAccessToken: elements["auth-token"].value.trim(),

        // infos
        buildCmd: buildCommands,          // []string
        exeCmd: executionCommands,        // []string
        env: environments,                // []string

        joinableForSeconds: joinableForSeconds ? parseInt(joinableForSeconds) : 0,
        maxConcurrentJoiners: maxConcurrentJoiners ? parseInt(maxConcurrentJoiners) : 0,
        timeoutSeconds: timeoutSeconds ? parseInt(timeoutSeconds) : 0,
        name: elements["name"].value.trim(),
        readme: elements["read-me"].value.trim(),
        keywords: keywords,
    };

    ALERT_MOD.call({
        isLoading: true,
        loadingText: "Loading..."
    });

    let promise = waitAtLeast(800, updateToaster(
        toaster.id,
        toaster.files,
        toaster.filepaths,
        toaster.gitURL,
        toaster.gitUsername,
        toaster.gitAccessToken,
        toaster.gitPassword,
        toaster.gitBranch,
        toaster.buildCmd,                   // []string
        toaster.exeCmd,                     // []string
        toaster.env,
        toaster.joinableForSeconds,         // int
        toaster.maxConcurrentJoiners,       // int
        toaster.timeoutSeconds,             // int
        toaster.name,
        toaster.readme,
        toaster.keywords,
    ));

    var handleBuildFn = function (cb) {
        if (!cb.success) {
            if (cb.build_error) {
                ALERT_MOD.call({
                    title: "Unsuccessful Build ",
                    text: `<pre class="buildlogs-pre">` + atob(cb.build_error) + "</pre>",
                    withCheckmark: false
                });
            } else {
                ALERT_MOD.call({
                    title: "Information",
                    text: cb.code + " : " + cb.message
                });
            }

            return;
        }

        if (gListToastersForSearch) {
            gListToastersForSearch = null;
        }

        if (imgToaster) {
            postToasterPicture(formDataImage, cb.toaster.id).then(cbImg => {
                if (cbImg && cbImg.success) {
                    ALERT_MOD.call({
                        title: "Information",
                        withCheckmark: true,
                        text: "Your Toaster has been successfully created.",
                        buttons: [
                            {
                                text: "See Logs",
                                onClick: function () {
                                    ALERT_MOD.close();

                                    var logs = "No logs";

                                    if (cb.build_logs && cb.build_logs.length > 0) {
                                        logs = `<pre class="buildlogs-pre">` + atob(cb.build_logs) + "</pre>";
                                    }

                                    ALERT_MOD.call({
                                        title: "Build Logs",
                                        withCheckmark: false,
                                        text: logs,
                                        buttons: [
                                            {
                                                text: "OK",
                                                onClick: function () {
                                                    ALERT_MOD.close();
                                                    window.location = "/toaster?id=" + cb.toaster.id;
                                                }
                                            },
                                        ]
                                    });
                                }
                            },
                            {
                                text: "OK",
                                onClick: function () {
                                    ALERT_MOD.close();
                                    window.location = "/toaster?id=" + cb.toaster.id;
                                }
                            },
                        ]
                    });
                }
                else if (cbImg && !cbImg.success) {
                    ALERT_MOD.call({
                        title: "Information",
                        text: "An error occurred during Toaster image upload. Please, try again or contact us.",
                        buttons: [
                            {
                                text: "OK",
                                onClick: function () {
                                    ALERT_MOD.close();
                                    window.location = "/toaster?id=" + cb.toaster.id;
                                }
                            },
                        ]
                    });
                }
            });
        }
        else {
            ALERT_MOD.call({
                title: "Information",
                withCheckmark: true,
                text: "Your Toaster has been successfully created.",
                buttons: [
                    {
                        text: "See Logs",
                        onClick: function () {
                            ALERT_MOD.close();

                            var logs = "No logs";

                            if (cb.build_logs && cb.build_logs.length > 0) {
                                logs = `<pre class="buildlogs-pre">` + atob(cb.build_logs) + "</pre>";
                            }

                            ALERT_MOD.call({
                                title: "Build Logs",
                                withCheckmark: false,
                                text: logs,
                                buttons: [
                                    {
                                        text: "OK",
                                        onClick: function () {
                                            ALERT_MOD.close();
                                            window.location = "/toaster?id=" + cb.toaster.id;
                                        }
                                    },
                                ]
                            });
                        }
                    },
                    {
                        text: "OK",
                        onClick: function () {
                            ALERT_MOD.close();
                            window.location = "/toaster?id=" + cb.toaster.id;
                        }
                    },
                ]
            });
        }

        localStorage.removeItem("markupValue");
    };

    promise.then((cb) => {
        if (cb && cb.success) {
            if (cb.build_id && cb.build_id.length > 0) {
                var build_id = cb.build_id;

                var checkRecursiveBuildend = function () {
                    getToasterBuildResult(build_id).then(cbbuild => {
                        if (cbbuild && cbbuild.success) {
                            if (!cbbuild.in_progress) {
                                clearInterval(checkbuildinterval);
                                handleBuildFn(cbbuild);
                                clearInterval(checkbuildinterval);
                            }
                        } else {
                            ALERT_MOD.call({
                                title: "Information",
                                text: cbbuild.code + " : " + cbbuild.message
                            });
                            clearInterval(checkbuildinterval);
                        }
                    });
                };

                var checkbuildinterval = setInterval(checkRecursiveBuildend, 5000);
            } else {
                handleBuildFn(cb);
            }
        } else {
            handleBuildFn(cb);
        }
    });

});

function toFilesAndFilesPath(filesArray) {
    let fileObjects = {
        filePaths: [],
        files: []
    };

    for (let i = 0; i < filesArray.length; i++) {
        let file = filesArray[i];

        if (file.type === "folder") {
            for (let j = 0; j < filesArray[i].files.length; j++) {
                let fileFromFolder = filesArray[i].files[j];

                fileObjects.files.push(fileFromFolder.file);
                fileObjects.filePaths.push(fileFromFolder.path);
            }
        }
        else if (file.type === "file") {
            fileObjects.files.push(file.file);
            fileObjects.filePaths.push(file.file.name);
        }
    }

    return fileObjects;
}
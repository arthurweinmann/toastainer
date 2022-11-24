import "local://includes/routes/toaster.js";
import "local://includes/counter.js";
import "local://includes/slider.js";
import "local://includes/toastTile.js";

var blockToasters = document.querySelector(".section-toasters");
var sectionStatsElem = document.querySelector(".section-stats");

getUsage((new Date().getMonth() + 1) + "", new Date().getFullYear() + "").then(cb => {
    if (cb && cb.success) {
        if (cb.usage) {
            let usage = cb.usage;

            console.log(usage);

            sectionStatsElem.classList.add("show");

            counterAnim("#stat1", 0, usage.runs ? usage.runs : 0, 300, "");
            counterAnim("#stat2", 0, usage.duration_ms ? usage.duration_ms : 0, 300, "MS");
            counterAnim("#stat3", 0, usage.cpu_seconds ? usage.cpu_seconds : 0, 300, "S");
            counterAnim("#stat4", 0, usage.ram_gbs ? usage.ram_gbs : 0, 300, "GBS");
            counterAnim("#stat5", 0, usage.net_ingress ? usage.net_ingress : 0, 300, "B");
            counterAnim("#stat6", 0, usage.net_egress ? usage.net_egress : 0, 300, "B");
        }
    }
});

handleDisplayToasters();

function handlePropagateLastToasters() {
    handleDisplayToasters();
}

function handleDisplayToasters() {
    handleRenderToasters(blockToasters, {
        sliderListClass: ".toaster-slider-1",
        isItemSlider: true,
        filterToasters: (toasters) => {
            return toasters.slice(-5).reverse();
        },
        deleteCallback: handlePropagateLastToasters,
        emptyCallback: () => {
            var sectionToastersElem = document.querySelector(".section-toasters");
            if (sectionToastersElem) { sectionToastersElem.style.display = "none"; }
        },
        endCallback: () => {
            var toasters1Slider = document.getElementById("toasters-1-slider");
            // center toasters slider
            toasters1Slider.scrollLeft = (toasters1Slider.scrollWidth - toasters1Slider.offsetWidth) / 2;
        }
    });

    handleRenderToasters(blockToasters, {
        sliderListClass: ".toaster-slider-2",
        isItemSlider: true,
        deleteCallback: handlePropagateLastToasters,
        filterToasters: (toasters) => {
            return toasters.slice(-10, -5).reverse();
        },
        emptyCallback: () => {
            var sectionToastersElem = document.querySelector(".section-toasters");
            if (sectionToastersElem) { sectionToastersElem.style.display = "none"; }
        },
        endCallback: () => {
            var toasters2Slider = document.getElementById("toasters-2-slider");
            // center toasters slider
            toasters2Slider.scrollLeft = ((toasters2Slider.scrollWidth - toasters2Slider.offsetWidth) / 2) - 75;
        }
    });
}

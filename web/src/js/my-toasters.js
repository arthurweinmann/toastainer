import "local://includes/routes/toaster.js";
import "local://includes/toastTile.js";

var blockToasters = document.querySelector(".section-toasters");

handleDisplayToasters();

function handleDisplayToasters() {
    handleRenderToasters(blockToasters, {
        emptyCallback: () => {            
            var emptyZone = document.querySelector(".empty__zone");
            if (emptyZone) { emptyZone.classList.add("show"); }
        }
    });

    /* todo later
     handleRenderToasters(blockToasters, {
        deleteCallback: handleDisplayToasters
    });
    */
}
import "local://includes/routes/user.js";

if (USER) {
    if (USER.active_billing) {
        var btnChooseBasicPlan = document.querySelector(".btn-choose-plan[data-plan='basic']");

        btnChooseBasicPlan.classList.add("is-active");
        btnChooseBasicPlan.classList.add("disabled");
        btnChooseBasicPlan.textContent = "Current plan"
    }
}

document.querySelectorAll(".btn-choose-plan").forEach(btnChoosePlan => {
    btnChoosePlan.addEventListener("click", function() {
        let typePlan = btnChoosePlan.getAttribute("data-plan");
        
        switch (typePlan) {
            case "basic":
                setupBilling();
                break;
            default:
        }
    })
});
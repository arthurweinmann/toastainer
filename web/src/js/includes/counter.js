function counterAnim(qSelector, start = 0, end, duration = 1000, type) {
    const target = document.querySelector(qSelector);
    let startTimestamp = null;
    const step = (timestamp) => {
        if (!startTimestamp) startTimestamp = timestamp;
        const progress = Math.min((timestamp - startTimestamp) / duration, 1);

        if (isInt(end)) {
            target.innerText = Math.floor(progress * (end - start) + start) + (type ? " " + type : "");
        }
        else {
            target.innerText = (progress * (end - start) + start).toFixed(2) + (type ? " " + type : "");        
        }

        if (progress < 1) {
            window.requestAnimationFrame(step);
        }
    };
    window.requestAnimationFrame(step);
};

function isInt(n) {
    return n % 1 === 0;
}
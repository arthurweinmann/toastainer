function waitAtLeast(time, promise) {
    const timeoutPromise = new Promise((resolve) => {
        setTimeout(resolve, time);
    });
    return Promise.all([promise, timeoutPromise]).then((values) => values[0]);
};
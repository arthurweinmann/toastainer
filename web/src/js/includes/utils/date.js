function timeSince(date) {
    var seconds = Math.floor(((new Date().getTime() / 1000) - date))

    var interval = seconds / 31536000;

    if (interval >= 1) {
        let intervalFloor = Math.floor(interval);
        return intervalFloor === 1 ? "Last year " : intervalFloor + " years ago";
    }
    interval = seconds / 2592000;
    if (interval >= 1) {
        let intervalFloor = Math.floor(interval);
        return intervalFloor === 1 ? "Last month" : intervalFloor + " months ago";
    }
    interval = seconds / 86400;
    if (interval >= 1) {
        let intervalFloor = Math.floor(interval);
        return intervalFloor === 1 ? "Yesterday" : intervalFloor + " days ago";
    }
    interval = seconds / 3600;
    if (interval >= 1) {
        let intervalFloor = Math.floor(interval);
        return intervalFloor === 1 ? "1 hour ago" : intervalFloor + " hours ago";
    }
    interval = seconds / 60;
    if (interval >= 1) {
        let intervalFloor = Math.floor(interval);
        return intervalFloor === 1 ? "1 minute ago" : intervalFloor + " minutes ago";
    }
    return Math.floor(seconds) + " seconds ago";
}

function getShortDate(date) {
    return date.toLocaleDateString('fr-FR');
}
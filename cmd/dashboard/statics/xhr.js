function get(uri) {
    return new Promise(function (resolve, reject) {
        let xhr = new XMLHttpRequest();
        xhr.onreadystatechange = function () {
            if (xhr.readyState !== 4) {
                return;
            }
            if (xhr.status >= 200 && xhr.status < 300) {
                if (xhr.getResponseHeader('content-type').indexOf('application/json') !== -1) {
                    let resp = JSON.parse(xhr.responseText);
                    resolve(resp);
                } else {
                    resolve(xhr.responseText);
                }
            } else {
                if (xhr.getResponseHeader('content-type').indexOf('application/json') !== -1) {
                    let resp = JSON.parse(xhr.responseText);
                    reject(resp)
                } else {
                    reject(xhr.responseText)
                }
            }
        };

        xhr.open('GET', uri, true);
        xhr.send();
    });
}
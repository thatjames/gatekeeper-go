export default Request = (method, url, data) => {
    return new Promise((resolve, reject) => {
        let req = new XMLHttpRequest();
        req.onreadystatechange = function () {
            if (this.readyState == 4) {
                if (this.status == 200) {
                    resolve(this.response);
                    return;
                } else if (this.status === 401 || this.status === 403) {
                    location.href = "/";
                } else {
                    reject(this.statusText);
                    return;
                }
            }
        };
        req.open(method, url);
        if (sessionStorage.getItem("token")){
            req.setRequestHeader("authorization", sessionStorage.getItem("token"));
        }
        if (data) {
            req.setRequestHeader("Content-Type", "application/json");
        }
        req.send(data);
        console.log(data);
    });
};
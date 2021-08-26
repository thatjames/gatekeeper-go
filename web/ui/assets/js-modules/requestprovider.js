export default Request = (method, url, data) => {
   return new Promise((resolve, reject) => {
        let req = new XMLHttpRequest();
        req.onreadystatechange = function() {
            if (this.readyState == 4 && this.status == 200) {
                resolve(this);
                console.log("Hello World");
            } else {
                console.log("Booooo!");
                console.log(this);
                reject(this);
            }
        };
        req.open(method, url);
        req.setRequestHeader("Content-Type", "application/json");
        req.send(data);
        console.log(data);
   });
};
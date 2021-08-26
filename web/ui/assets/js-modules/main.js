import Request from './requestprovider.js';
import { GetFormData } from './utils.js';

window.Login = function() {
    var form = document.getElementById("login-form");
    var data = GetFormData(new FormData(form));
    Request("POST", "http://localhost:5000/api/login", JSON.stringify(data)).then((res) =>{
        window.location = "/main";
        console.log("done and passed");
    }).catch((err) => {
        console.log(err);
    });
    console.log("done");
};
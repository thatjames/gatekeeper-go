import Request from './requestprovider.js';
import { GetFormData } from './utils.js';

export function Login(e) {
    e.preventDefault();
    var form = document.getElementById("login-form");
    var data = GetFormData(new FormData(form));
    Request("POST", "api/login", JSON.stringify(data)).then((res) =>{
        location.href = "/main";
        console.log("done");
    }).catch((err) => {
        console.log("Failed: " + err);
    });
}
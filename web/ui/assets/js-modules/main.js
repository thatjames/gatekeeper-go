import Request from './requestprovider.js';
import { GetFormData } from './utils.js';

export function Login(e) {
    e.preventDefault();
    var form = document.getElementById("login-form");
    var data = GetFormData(new FormData(form));
    Request("POST", "api/login", JSON.stringify(data)).then((res) =>{
        sessionStorage.setItem("token", "set");
        location.href = "/main";
    }).catch((err) => {
        console.log("Failed: " + err);
    });
}

export function GetPage(e) {
    e.preventDefault();
    Request("GET", "page/"+e.target.id).then((res)=>{
        document.getElementById("content").innerHTML = res;
    }).catch((err) => {
        console.log("error: " + err);
    });
}
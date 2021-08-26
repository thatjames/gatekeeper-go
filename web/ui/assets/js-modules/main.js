import Request from './requestprovider.js';
import { GetFormData } from './utils.js';

window.Login = function() {
    var form = document.getElementById("login-form");
    var data = GetFormData(new FormData(form));
    Request("POST", "/api/login", data);
};
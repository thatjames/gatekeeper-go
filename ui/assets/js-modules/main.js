import Request from './requestprovider.js';

window.Login = function() {
    var form = document.getElementById("login-form");
    var data = getFormData(new FormData(form));
    console.log(data);
    Request("yes");
};

function getFormData(form){
    var result = {};
    form.forEach((value, key) => result[key] = value);    
    return result;
}
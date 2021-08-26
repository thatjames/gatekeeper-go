export function GetFormData(form){
    var result = {};
    form.forEach((value, key) => result[key] = value);    
    return result;
}
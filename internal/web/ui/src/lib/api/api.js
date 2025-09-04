const environments = {
    dev: {
        name: "dev",
        url: "http://localhost:8085/api/"
    },
    prod: {
        name: "prod",
        url: "/api/"
    }
}

let environment = environments.dev

export const get = async (path) => {
    return await fetch(environment.url + path)
}

export const post = async (path, data) => {
    return fetch(environment.url + path, {
        method: "POST",
        body: JSON.stringify(data)
    })
}
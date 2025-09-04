const environments = {
  dev: {
    url: "http://localhost:8085/api/",
  },
  live: {
    url: "/api/",
  },
};

let env = import.meta.env.PROD ? environments.live : environments.dev;

export class API {
  networkRequest(method, path, data) {
    let url = env.url + path;
    return fetch(url, {
      method: method,
      body: data ? JSON.stringify(data) : null,
    });
  }

  get(path) {
    return this.networkRequest("get", path);
  }

  post(path, data) {
    return this.networkRequest("post", path, data);
  }

  put(path, data) {
    return this.networkRequest("put", path, data);
  }

  delete(path, data) {
    return this.networkRequest("delete", path, data);
  }
}

export const api = new API();

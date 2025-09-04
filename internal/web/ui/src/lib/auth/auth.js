import { post } from "$lib/api/api"

export const login = (username, password) => {
    post("auth/login", { username: username, password: password})
} 
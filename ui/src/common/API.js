import { addToast } from './toasts.js';

export async function callAPI(endpoint, parameters) {
    let promise = new Promise((resolve, reject) => {
        fetch(API+endpoint, parameters)
        .then(response => {
            if (response.status == 204) {
                return {}
            }
            return response.json()
        })
        .then(data => {
            if (data.error) {
                addToast({
                    type: "error",
                    title: "Error!",
                    message: data.error,
                })
                return;
            }
            resolve(data);
        })
        .catch((error) => {
            console.error(error);
            addToast({
                type: "error",
                title: "Error!",
                message: error.toString(),
            })
            reject(error);
        });
    });

    return promise
}
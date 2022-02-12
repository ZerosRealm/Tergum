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
            let errorMsg = data.error;
            if (data.message != undefined && data.message != "") {
                errorMsg = data.message;
            }

            if (errorMsg != undefined) {
                addToast({
                    type: "error",
                    title: "Error!",
                    message: errorMsg,
                })
                return;
            }
            resolve(data);
        })
        .catch((error) => {
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
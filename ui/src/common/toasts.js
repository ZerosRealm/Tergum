import { writable } from 'svelte/store';

export const toasts = writable([]);

export const addToast = (toast) => {
    toasts.update(toasts => toasts = [...toasts, toast]);
}

export default {
	addToast
}
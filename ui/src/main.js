import './process.js';
import App from './App.svelte';

console.log("API endpoint:", API)

const app = new App({
	target: document.body,
	props: {}
});

export default app;
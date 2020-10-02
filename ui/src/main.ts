import App from "./App.svelte";
import { login } from "./api";

declare global {
	interface Window {
    gapi: any;
	}
}

const app = new App({
	target: document.body,
});

export default app;
